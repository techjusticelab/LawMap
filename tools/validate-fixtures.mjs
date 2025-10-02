#!/usr/bin/env node
import fs from 'node:fs';
import readline from 'node:readline';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import Ajv from 'ajv';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const schemaPath = path.resolve(__dirname, '..', 'API', 'docs', 'schemas', 'graph_item.schema.json');
const baseDir = path.dirname(schemaPath);

function loadSchema(file) {
  const p = path.resolve(baseDir, file);
  return JSON.parse(fs.readFileSync(p, 'utf8'));
}

const schemas = {
  'graph_item.schema.json': loadSchema('graph_item.schema.json'),
  'graph_node.schema.json': loadSchema('graph_node.schema.json'),
  'graph_edge.schema.json': loadSchema('graph_edge.schema.json'),
};

const ajv = new Ajv({ strict: false, allErrors: true });
for (const [id, schema] of Object.entries(schemas)) {
  ajv.addSchema(schema, id);
}
const validate = ajv.getSchema('graph_item.schema.json');

const examplesFile = path.resolve(__dirname, '..', 'API', 'docs', 'EXAMPLES.graph.jsonl');
if (!fs.existsSync(examplesFile)) {
  console.error(`Missing examples file: ${examplesFile}`);
  process.exit(1);
}

let lineNum = 0;
let errors = 0;
const rl = readline.createInterface({ input: fs.createReadStream(examplesFile), crlfDelay: Infinity });
for await (const line of rl) {
  lineNum++;
  const trimmed = line.trim();
  if (!trimmed) continue;
  let obj;
  try {
    obj = JSON.parse(trimmed);
  } catch (e) {
    console.error(`Line ${lineNum}: invalid JSON`);
    errors++;
    continue;
  }
  const ok = validate(obj);
  if (!ok) {
    console.error(`Line ${lineNum}: schema errors\n`, validate.errors);
    errors++;
  }
}

if (errors > 0) {
  console.error(`Validation failed with ${errors} error(s).`);
  process.exit(1);
}
console.log('Fixture validation passed.');

# Primary Law Sources for California Public Defense

This repository lists official, non-API data sources you can browse or download for the authorities PDs rely on. Use these as canonical ingestion points for your search/indexing pipeline.

---

# Federal

## United States Constitution
- Text and annotations (Constitution Annotated): https://www.govinfo.gov/app/collection/GPO-CONAN
- National Archives (founding document): https://www.archives.gov/founding-docs/constitution

## US Code (federal statutes)
- Office of the Law Revision Counsel bulk/downloads: https://uscode.house.gov/download/download.shtml
- Authenticated PDFs at GovInfo: https://www.govinfo.gov/app/collection/USCODE

## Code of Federal Regulations (annual editions)
- GovInfo collection: https://www.govinfo.gov/app/collection/CFR

## eCFR (current, continuously updated regs)
- Browse: https://www.ecfr.gov/

## Federal Register (rules, notices)
- Browse/search: https://www.federalregister.gov/

## Federal Case Law (opinions)
- U.S. Supreme Court opinions: https://www.supremecourt.gov/opinions/opinions.aspx
- Ninth Circuit opinions: https://www.ca9.uscourts.gov/opinions/
- Aggregated/open access (browse and bulk downloads available): https://www.courtlistener.com/

## Federal Sentencing
- U.S. Sentencing Guidelines Manual and materials: https://www.ussc.gov/guidelines

## Federal Dockets and Filings
- PACER portal (paid): https://pacer.uscourts.gov
- RECAP archive (public mirrors of PACER filings): https://www.courtlistener.com/recap/

---

# California

## Cal Laws (codes + constitution)
Canonical site: https://leginfo.legislature.ca.gov/
- Homepage “Downloadable Database” link (bulk statutory text and legislative data)

### California Constitution — CONS
- Source: https://leginfo.legislature.ca.gov/ (California Law → Constitution)

### California Codes (statutes)
- Business and Professions Code — BPC
- Civil Code — CIV
- Code of Civil Procedure — CCP
- Commercial Code — COM
- Corporations Code — CORP
- Education Code — EDC
- Elections Code — ELEC
- Evidence Code — EVID
- Family Code — FAM
- Financial Code — FIN
- Fish and Game Code — FGC
- Food and Agricultural Code — FAC
- Government Code — GOV
- Harbors and Navigation Code — HNC
- Health and Safety Code — HSC
- Insurance Code — INS
- Labor Code — LAB
- Military and Veterans Code — MVC
- Penal Code — PEN
- Probate Code — PROB
- Public Contract Code — PCC
- Public Resources Code — PRC
- Public Utilities Code — PUC
- Revenue and Taxation Code — RTC
- Streets and Highways Code — SHC
- Unemployment Insurance Code — UIC
- Vehicle Code — VEH
- Water Code — WAT
- Welfare and Institutions Code — WIC
- Source for all codes above: https://leginfo.legislature.ca.gov/ (California Law tab; “Quick Code Search” lists each code)

### Historical/Legacy Legislative Materials
- Measures prior to 1999 archive: http://leginfo.ca.gov

## California Rules of Court
- Judicial Council (HTML/PDF): https://www.courts.ca.gov/rules-forms/rules-court

## Local Rules (County Superior Courts)
- Los Angeles Superior Court: https://www.lacourt.org
- San Francisco Superior Court: https://www.sfsuperiorcourt.org
- San Diego Superior Court: https://www.sdcourt.ca.gov
- Sacramento Superior Court: https://www.saccourt.ca.gov
- Note: Each court posts local rules as HTML/PDF on its own site.

## Jury Instructions
- CALCRIM (Criminal) — official PDFs
- CACI (Civil) — official PDFs
- Judicial Council Publications page: https://www.courts.ca.gov/partners/317.htm

## California Case Law (opinions)
- California Courts opinions portal: https://www.courts.ca.gov/opinions.htm
- Aggregated/open access with metadata (browse/bulk): https://www.courtlistener.com/

## California Regulations
- California Code of Regulations (CCR) browsing site (hosted via Westlaw): https://govt.westlaw.com/calregs/
- Office of Administrative Law publications: https://oal.ca.gov/publications/
- Agency policy/regulatory materials (often relevant to credits, DMV, etc.):
  - CDCR: https://www.cdcr.ca.gov
  - DMV: https://www.dmv.ca.gov

## Legislative Materials (bills, histories)
- California Legislative Information (bills, histories, analyses): https://leginfo.legislature.ca.gov/ (Bill Information tab)
- “Downloadable Database” link on the LegInfo homepage for bulk downloads

---

# What’s Inside Each (PD-focused quick reference)

- Penal Code (PEN): Crimes, elements, defenses, procedures (e.g., §1538.5 suppression), sentencing provisions, enhancements.
- Evidence Code (EVID): Admissibility, privileges, hearsay exceptions, character evidence.
- Vehicle Code (VEH): DUI, license, traffic offenses and consequences.
- Health & Safety (HSC): Controlled substances, public health-related offenses.
- Welfare & Institutions (WIC): Juvenile proceedings, dependency/mental health interfaces.
- Code of Civil Procedure (CCP): Subpoenas, discovery in limited contexts, time computation.
- Government Code (GOV): Public records, claims statutes, PRA; some criminal justice admin provisions.
- Rules of Court (CRC): Filing/formatting, deadlines, appellate rules, sentencing rules interplay.
- CALCRIM: Model jury instructions with elements/bench notes (authorities cited).
- CA/Federal Constitutions: Rights affecting searches, interrogations, counsel, confrontation.
- CA/Federal Case Law: Binding/persuasive interpretations of the above.
- CFR/eCFR & Federal Register: Federal administrative rules (collateral consequences).
- CCR & Agency Policies: State administrative rules (custody credits, DMV sanctions).

---

# Notes
- LegInfo states: Information made available on the site is in the public domain (Gov. Code §10248.5).
- CCR has contractual restrictions on bulk use from the official host; consider agency PDFs or licensed sources for comprehensive ingestion.
- For ingestion, prefer official bulk “Downloadable Database” (LegInfo), US Code bulk XML, and GovInfo collections. Respect each site’s terms when scraping.

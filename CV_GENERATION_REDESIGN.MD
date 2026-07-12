# CV Generation Template Redesign — Implementation Plan

## Architecture

CV PDF is generated **server-side in Go** using `gofpdf` (`internal/service/pdf_service.go`).
Frontend (`cv/page.tsx`) only triggers generation via API and displays the result.

This plan modifies **only** the Go PDF drawing code. No backend logic, data retrieval, or frontend code changes.

---

## Current Layout (3 Pages)

| Page | Content | Function |
|------|---------|----------|
| 1 | Two-column profile sheet: Photo + Name + Application Details (left), Header + Personal Info + Passport + Skills + Exp Abroad + Languages (right) | `drawApplicationProfileSheet` |
| 2 | Raw passport document image | `appendPassportSection` |
| 3 | Full body photo | `appendFullBodyPhotoSection` |

**Dead code (defined but never called):** left untouched.

---

## Color System

| Role | RGB | Hex | Usage |
|------|-----|-----|-------|
| Primary Navy | `(27, 42, 74)` | `#1B2A4A` | Header bars, name banner |
| Deep Royal Blue | `(37, 99, 235)` | `#2563EB` | Accent lines, links |
| Card Header Blue | `(30, 58, 138)` | `#1E3A8A` | Section header text |
| Accent Emerald | `(5, 150, 105)` | `#059669` | Skill badges, language proficiency |
| Accent Amber | `(217, 119, 6)` | `#D97706` | Intermediate language proficiency |
| Accent Red | `(220, 38, 38)` | `#DC2626` | Expiring passport, alerts |
| Page Background | `(248, 250, 252)` | `#F8FAFC` | Full-page subtle fill |
| Card Background | `(255, 255, 255)` | `#FFFFFF` | All cards |
| Border Light | `(226, 232, 240)` | `#E2E8F0` | Card borders, dividers |
| Text Muted | `(100, 116, 139)` | `#64748B` | Labels |
| Text Dark | `(30, 41, 59)` | `#1E293B` | Values |
| Placeholder Gray | `(241, 245, 249)` | `#F1F5F9` | Photo placeholder background |

---

## Layout Constants

| Constant | Current | New |
|----------|---------|-----|
| Page margins | 15mm | **18mm** |
| Left column X | 12mm | **18mm** |
| Left column width | 65mm | **58mm** |
| Right column X | 82mm | **88mm** |
| Right column width | 115mm | **104mm** |
| Gap between columns | 5mm | **12mm** |
| Content width | 186mm | **174mm** |
| A4 page width | 210mm | 210mm |
| A4 page height | 297mm | 297mm |
| Footer height | — | **12mm** |
| Auto page break margin | disabled | **25mm from bottom** |

---

## Overflow Handling Rules (Applied to ALL Steps)

Every text value drawn in the redesign must follow these rules, in order of priority:

| Rule | Implementation |
|------|---------------|
| 1. Empty → fallback | `compactValue(value, "N/A")` before any draw |
| 2. Too wide → shrink | `drawFitPDFText` shrinks font from baseSize down to minSize |
| 3. Still too wide → truncate | `ellipsizePDFText` appends "..." after shrinking |
| 4. Multi-line → wrap | `MultiCell` for long names, paragraphs |
| 5. Image fails → placeholder | `drawPhotoPlaceholder` draws grey rect with "PHOTO" text |

**Existing helpers reused (NOT rewritten):**
- `addImageFromBytesFit` — aspect-ratio-preserving image placement
- `drawFitPDFText` — font shrinking until text fits
- `ellipsizePDFText` — truncation with "..."
- `compactValue` — empty string fallback

---

## Card Height Rules (Applied to ALL Steps)

Every card uses a **formula** for height, never a raw constant:

| Card | Height Formula | Type |
|------|---------------|------|
| Branding Header | `14mm` | Fixed (single line) |
| Photo Panel | `76mm` | Fixed (standard photo area) |
| Name Banner | `min(14 + calculatedTextHeight, 36)` | **Dynamic** |
| Metric Chips | `10mm` | Fixed |
| Application Details | `14 + len(rows) * 7.5` | **Dynamic** |
| Personal Information | `13 + 5 * 7` | Fixed grid (5 rows × 7mm) |
| Passport Details | `13 + 2 * 9` | Fixed grid (2 rows × 9mm) |
| Work Experience | `13 + ceil(len(skills)/2) * 8` | **Dynamic** |
| Experienced Abroad | `20 + len(rows) * 7` | **Dynamic** ← already dynamic |
| Languages | `20 + len(langEntries) * 7` | **Dynamic** ← already dynamic |

**Positioning rule:** `currentRY` is always updated with `+= cardHeight + gap` so the next section anchors to the previous section's bottom.

---

## Implementation Plan (13 Increments)

---

### Step 0 — Utility Infrastructure

**Files:** `internal/service/pdf_service.go`

**Add new package-level constants:**
- `PageW = 210.0`, `PageH = 297.0`, `Margin = 18.0`, `ContentW = 174.0`
- Color triples for every color in the palette (named: `navyR, navyG, navyB`, etc.)
- `CardRadius = 4.0`

**Add new helper functions:**

```go
// drawCard — white rounded rect with #E2E8F0 border
func drawCard(pdf *gofpdf.Fpdf, x, y, w, h float64)

// drawCardHeader — colored header band inside a card
func drawCardHeader(pdf *gofpdf.Fpdf, x, y, w, h float64,
    iconName, title string, bgR, bgG, bgB int, textR, textG, textB int)

// drawLabelValue — muted label + bold value using drawFitPDFText for overflow
// Internally calls drawFitPDFText with baseSize, minSize for the value
func drawLabelValue(pdf *gofpdf.Fpdf, x, y, labelW, valW float64,
    label, value string, valBaseSize, valMinSize float64) float64

// drawPageFooter — thin line + generation date + page number at page bottom
func drawPageFooter(pdf *gofpdf.Fpdf)

// drawPhotoPlaceholder — grey #F1F5F9 rect with "PHOTO" text when image unavailable
func drawPhotoPlaceholder(pdf *gofpdf.Fpdf, x, y, w, h float64)

// drawTableHeaderRow — dark nav row with white column headers
func drawTableHeaderRow(pdf *gofpdf.Fpdf, x, y, w, h float64,
    columns []string, colWidths []float64)

// drawTableDataRow — alternating row with label/value cells
func drawTableDataRow(pdf *gofpdf.Fpdf, x, y float64,
    cells []string, colWidths []float64, rowIndex int, isLast bool) float64

// drawBadgePill — colored rounded rect with centered white text
func drawBadgePill(pdf *gofpdf.Fpdf, x, y, w, h float64,
    text string, bgR, bgG, bgB int)
```

**Risk:** None — all new code, no existing callers modified.

---

### Step 1 — Page Layout, Margins & Auto Page Break

**Function:** `GenerateCandidateCV` (line 73)

**Changes:**
- `pdf.SetMargins(Margin, Margin, Margin)` (was 15)
- Enable auto-page-break: `pdf.SetAutoPageBreak(true, 25)` (was `false, 0`)
  - This prevents content from spilling off the page
  - Sets bottom margin of 25mm so footer has room
- After `pdf.AddPage()`, draw full `#F8FAFC` background fill covering the entire page
- After each page's main content, call `drawPageFooter`

**Overflow handling:**
- With auto-page-break enabled, if page 1 right column grows too tall, content flows to a new page 2 (ahead of the passport/photo pages)
- The plan must ensure typical profiles fit page 1. Worst case: if too many skills/languages cause overflow, auto-page-break handles it gracefully
- Total available height: `PageH - 2*Margin - footerMargin` = `297 - 36 - 25 = 236mm`
- Estimated right column max height: branding(26) + personal(48) + passport(32) + skills(25-80) + exp(dynamic) + lang(dynamic) + gaps(~30) ≤ ~236mm ✓

**Risk:** Medium — verify all content still fits on page 1 for standard profiles.

---

### Step 2 — Branding Header

**Function:** `drawApplicationProfileSheet` lines 247-264

**Replace with:**

```
y=12:  ┌────────────────────────────────────────┐
       │ Ethiopian Agency    [LOGO]    Foreign  │  ← bar, 14mm tall, r=3
       └────────────────────────────────────────┘
y=28:  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  ← #2563EB accent line, 0.4mm
y=32:         CANDIDATE APPLICATION PROFILE       ← 15pt bold #1B2A4A centered
y=42:       Prepared by Ethiopian Agency          ← 8pt #64748B centered
y=48:  (right column content starts here)
```

- Header bar at `(Margin, 12)` size `(ContentW, 14)`
  - `#F8FAFC` fill, `#E2E8F0` 1px border, r=3
  - Left third: Ethiopian agency name → `drawSafeValue` max width 55mm, base 9pt, min 6pt
  - Center: logo image 34×10mm if available, vertically centered
  - Right third: Foreign agency name → `drawSafeValue` max width 55mm, base 9pt, min 6pt
- Accent line at `y=28` across full content width
- Title "CANDIDATE APPLICATION PROFILE" at `(Margin, 33)` 15pt bold `#1B2A4A` centered
- Subtitle at `(Margin, 43)` 8pt `#64748B` centered
- Set `currentRY = 50` for right column

**Overflow handling:**
- `drawSafeValue` shrinks agency names down to 6pt if >55mm wide
- If still too wide after shrinking, use the default fallback name

---

### Step 3 — Candidate Photo Panel

**Function:** `drawApplicationProfileSheet` lines 266-296

**Replace with:**

```
y=50:  ┌─────────────────────┐
       │                     │
       │     PHOTO AREA      │  ← white card 58×76mm
       │     58×76mm         │
       │                     │
       │ ┌──┐ ─────────────  │  ← name banner overlays bottom 24-36mm
       │ │  │ FIRST_NAME     │  ← #1B2A4A bg, #2563EB left stripe
       │ │  │ LAST_NAME      │
       │ └──┘ ─────────────  │
       └─────────────────────┘
y=130: ┌─────────────────────┐
       │  28  │  5YRS  │  3  │  ← metric chips, 58×10mm
       └─────────────────────┘
y=144: (Application Details starts here)
```

- Photo card at `(Margin, 50)` size `(58, 76)` — white fill, `#E2E8F0` border, r=4
- Photo image: 2mm padding inside card, aspect-ratio preserved via `addImageFromBytesFit`
  - If image fails or is nil → call `drawPhotoPlaceholder` inside the card instead
- Name banner: overlays bottom of photo card
  - Banner height: `max(24, 14 + multiLineNameHeight(name, 54mm, 12pt))` capped at `36mm`
  - Banner Y: `50 + 76 - bannerHeight`
  - `#1B2A4A` fill
  - Left 3mm accent stripe in `#2563EB`
  - First name split from last via `strings.Fields`, each drawn centered with `drawFitPDFText` base 12pt, min 8pt
  - If name doesn't split, draw full name on one line
- Metric chips at `(Margin, 130)` size `(58, 10)`
  - Three equal chips: Age | Experience | Languages
  - Each: `#F1F5F9` fill, `#E2E8F0` border, r=3, `#1E293B` 7pt bold centered text
  - Values via `formatIntPointer` / `formatExperienceLevel`
- Set `appY = 144` for Application Details

**Overflow handling:**
- Name text shrinks to min 8pt if very long
- Banner caps at 36mm (3 lines) to avoid taking too much vertical space
- If photo image fails → grey placeholder with "PHOTO" label so the area is not blank

---

### Step 4 — Application Details Card

**Function:** `drawApplicationProfileSheet` lines 298-343

**Replace with:**

```
y=144: ┌──────────────────────┐
       │ APPLICATION DETAILS   │  ← navy header band 9mm
       ├──────────────────────┤
       │ Applied For   HOUSEMAID│  ← 7.5mm per row
       │ Ref. No       ABC12345│  ← label 7pt #64748B | value 7.5pt bold
       │ Contract      2 YRS   │
       │ Salary        $400    │
       │ ...                   │
       │ Date          8/7/2026│
       └──────────────────────┘
```

- Card at `(Margin, appY)` width `58mm`, height `HeaderH(9) + len(appRows) * RowH(7.5)`
- White fill, `#E2E8F0` 1px border, r=4
- Header: `#1B2A4A` fill, 9mm, white 8pt bold "APPLICATION DETAILS" centered
- 9 rows, each 7.5mm:
  - Label: `x+4`, width 28mm, `#64748B` 7pt left
  - Value: `x+32`, width 22mm, `#1E293B` 7.5pt bold, right-aligned via `drawSafeValue`
  - Divider `#F1F5F9` line after each row (except last)
- Row data (unchanged):
  - Applied For, Ref. No, Contract Period, Monthly Salary, Applied Contract, Experience, Bio. Met. ID, B. Met Rg./S., Date

**Overflow handling:**
- Values use `drawSafeValue` with base 7.5pt, min 5.5pt
- Long salary strings, country names, or reference numbers shrink to fit

---

### Step 5 — Personal Information Card

**Function:** `drawApplicationProfileSheet` lines 348-392

**Replace with:**

```
┌──────────────────────────────────────────────┐
│ 👤 PERSONAL INFORMATION                       │  ← #F1F5F9 header 9mm
├────────────────┬─────────────────────────────┤
│ 🚩 Nationality  │ ⛪ Religion                 │  ← each row 7mm
│    ETHIOPIAN    │    CHRISTIAN               │
│ 📅 Date of Birth│ ❤️ Marital Status          │
│    01-Jan-95    │    MARRIED                 │
│ 👤 Age          │ 👶 Children                │
│    30 YRS       │    2                      │
│ 📍 Place of Birth│ 🎓 Education Level        │
│    ADDIS ABABA  │    HIGH SCHOOL             │
│ ♀️ Gender       │                           │
│    FEMALE       │                           │
└────────────────┴─────────────────────────────┘
```

- Card at `(88, 50)` width `104mm`, height `9 + 5*7 = 44mm`
- White fill, `#E2E8F0` 1px border, r=4
- Header: `#F1F5F9` band 9mm, user.png icon (5×5mm) + "PERSONAL INFORMATION" 9pt bold `#1E3A8A`
- Content: 5 rows × 7mm, each row split into two fields:
  - Left column (0-48mm): label 7pt `#64748B` | value 7.5pt bold `#1E293B` via `drawSafeValue`
  - Right column (52-100mm): same pattern
  - Separator: thin `#E2E8F0` line at 50mm
- Fields: Nationality, DOB, Age, Place of Birth, Gender (left); Religion, Marital Status, Children, Education Level (right)
  - Last row: right column blank (only 4 fields on right vs 5 on left)
- `currentRY += 48`

**Overflow handling:**
- Each value uses `drawSafeValue` base 7.5pt, min 5.5pt, max width 44mm
- All values can hold typical content ("UNITED KINGDOM" at 7.5pt is ~40mm → fits)
- Empty values: `compactValue(value, "N/A")` applied before draw

---

### Step 6 — Passport Details Card

**Function:** `drawApplicationProfileSheet` lines 395-453

**Replace with:**

```
┌──────────────────────────────────────────────┐
│ 🛂 PASSPORT DETAILS                           │  ← #F1F5F9 header 9mm
├─────────────────────┬────────────────────────┤
│ Passport No.        │ Expiry Date            │  ← row 1: 9mm
│   AB123456          │   05-Jan-30            │
│ Issue Date          │ Remaining Years        │  ← row 2: 9mm
│   05-Jan-20         │   4 years              │
└─────────────────────┴────────────────────────┘
```

- Card at `(88, currentRY)` width `104mm`, height `9 + 2*9 = 27mm`
- White fill, `#E2E8F0` 1px border, r=4
- Header: `#F1F5F9` band 9mm, passport.png icon + "PASSPORT DETAILS" 9pt bold
- Content: 2 rows × 9mm, each row has 2 fields
  - Row 1: "Passport No." (left) | "Expiry Date" (right)
  - Row 2: "Issue Date" (left) | "Remaining Years" (right)
  - Each field: label 7pt `#64748B` at top of cell, value 9pt bold `#1E293B` below
  - Center divider: `#E2E8F0` line at 52mm
  - If expiry is soon: value in `#DC2626` (red)
- Passport data resolved from `passportData` first, falls back to `candidate.*` fields
- `currentRY += 31`

**Overflow handling:**
- Passport number uses `drawSafeValue` base 9pt, min 6pt, max width 46mm
- "Remaining Years" text (e.g. "3 years 2 months") fits 46mm at 9pt
- Expired: shown as "EXPIRED" in red bold

---

### Step 7 — Work Experience (Domestic) Card

**Function:** `drawApplicationProfileSheet` lines 455-515

**Replace with:**

```
┌──────────────────────────────────────────────┐
│ 💼 WORK EXPERIENCE (DOMESTIC)                 │  ← #F1F5F9 header 9mm
├──────────────────────────────────────────────┤
│ ✓ CLEANING        ✓ COOKING                  │  ← skill badges, 8mm per row
│ ✓ BABY SITTING    ✓ LAUNDRY                  │  ← each badge: #ECFDF5 fill
│ ✓ ELDERLY CARE                               │  ← #059669 text, r=3
└──────────────────────────────────────────────┘
```

- Card at `(88, currentRY)` width `104mm`, height `9 + 4 + ceil(len(skills)/2) * 8 + 4`
  - 4mm top padding + rows + 4mm bottom padding
- Header: briefcase.png icon + "WORK EXPERIENCE (DOMESTIC)" 9pt bold
- Skills drawn as badge tags in a 2-column grid:
  - Each badge: `x + 4`, width `46mm`, height `7mm`
  - `#ECFDF5` fill, `#059669` 0.5px border, r=3
  - Checkmark icon (4×4mm) at left
  - Skill name: `drawSafeValue` base 7pt, min 5pt, max width 36mm, `#059669` bold
  - Gap between columns: 4mm
  - Row height: 8mm
  - Top padding: 4mm
- Empty state: "No domestic work experience recorded" — `#64748B` 7.5pt centered
- `currentRY += cardHeight + 6`

**Overflow handling:**
- Long skill names shrink to min 5pt
- If a skill name is extremely long (>36mm at 5pt), truncates with "..."
- Empty skills: show message instead of blank card body

---

### Step 8 — Experienced Abroad Table

**Function:** `drawApplicationProfileSheet` lines 517-582

**Replace with:**

```
┌──────────────────────────────────────────────┐
│ 🌍 EXPERIENCED ABROAD                         │  ← #F1F5F9 header 9mm
├──────┬────────────────────────┬───────────────┤
│  #   │ COUNTRY                │ DURATION      │  ← #1B2A4A header row 6mm
├──────┼────────────────────────┼───────────────┤
│  1   │ SAUDI ARABIA           │     3 YEARS   │  ← 7mm per row
│  2   │ UNITED ARAB EMIRATES   │     2 YEARS   │  ← alternating bg
└──────┴────────────────────────┴───────────────┘
```

- Card at `(88, currentRY)` width `104mm`, height `9 + 4 + 6 + len(rows)*7 + 4`
  - 4mm top pad + header row + data rows + 4mm bottom pad
- Header: globe.png icon + "EXPERIENCED ABROAD" 9pt bold
- Table header: `#1B2A4A` fill, 6mm
  - Columns: **#** (12mm, center) | **COUNTRY** (48mm, left) | **DURATION** (40mm, right)
  - White 7pt bold text
- Data rows: 7mm each, alternating `#F8FAFC` / `#FFFFFF`
  - #: `#64748B` 7pt centered
  - Country: `#1E293B` 7.5pt bold left via `drawSafeValue`
  - Duration: `#059669` 7.5pt bold right via `drawSafeValue`
- Empty state: "FIRST TIME" badge, 36×8mm, `#DBEAFE` fill, `#2563EB` 8pt bold centered
- `currentRY += cardHeight + 6`

**⚠️ Already dynamic in current code.** The formula `18 + len(expRows)*7` already adjusts height. This step restyles the visual only.

**Overflow handling:**
- Country name uses `drawSafeValue` base 7.5pt, min 5.5pt, max width 44mm
- "SAUDI ARABIA" at 7.5pt ≈ 30mm → fits easily
- Very long country names shrink or truncate
- No rows → "FIRST TIME" badge centered

---

### Step 9 — Language Table

**Function:** `drawApplicationProfileSheet` lines 584-654

**Replace with:**

```
┌──────────────────────────────────────────────┐
│ 💬 LANGUAGES                                  │  ← #F1F5F9 header 9mm
├────────────────────────┬─────────────────────┤
│ LANGUAGE               │ PROFICIENCY          │  ← #1B2A4A header 6mm
├────────────────────────┼─────────────────────┤
│ 🗨️ AMHARIC            │     [FLUENT]        │  ← green pill
│ 🗨️ ENGLISH            │  [INTERMEDIATE]      │  ← amber pill
│ 🗨️ ARABIC             │     [BASIC]         │  ← blue pill
└────────────────────────┴─────────────────────┘
```

- Card at `(88, currentRY)` width `104mm`, height `9 + 4 + 6 + len(langEntries)*7 + 4`
- Header: chat.png icon + "LANGUAGES" 9pt bold
- Table header: `#1B2A4A` fill, 6mm
  - Columns: **LANGUAGE** (56mm, left) | **PROFICIENCY** (44mm, right)
  - White 7pt bold text
- Data rows: 7mm each, alternating `#F8FAFC` / `#FFFFFF`
  - Language: chat icon 4×4mm + `drawSafeValue` base 8pt, min 5.5pt, max width 48mm, `#1E293B` bold
  - Proficiency: badge pill 26×6mm, r=3, colored bg:
    - Fluent/Advanced/Native: `#059669`
    - Intermediate: `#D97706`
    - Basic/default: `#2563EB`
  - White 7pt bold text centered in pill
- Empty state: "FIRST TIME" badge (same pattern as experience abroad)
- `currentRY += cardHeight + 6`

**⚠️ Already dynamic in current code.** The formula `18 + len(langEntries)*7` already adjusts height. This step restyles the visual only.

**Overflow handling:**
- Language name uses `drawSafeValue` base 8pt, min 5.5pt, max width 48mm
- Proficiency text (max ~9 chars like "INTERMEDIATE") fits in 26mm pill at 7pt
- Long language names shrink or truncate in the badge area

---

### Step 10 — Passport Attachment Page (Page 2)

**Functions:** `appendPassportSection` (line 1156), `drawAttachmentImageFrame` (line 1180), `drawAttachmentHeader` (line 1170)

**Rewrite `drawAttachmentHeader`:**
- Navy `#1B2A4A` bar at `(Margin, 12)` size `(ContentW, 14)`, r=3
- White "ATTACHMENT — PASSPORT DOCUMENT" 11pt bold centered

**Rewrite `drawAttachmentImageFrame`:**
- Card frame at `(Margin, 34)` size `(ContentW, 225)`:
  - `#F8FAFC` fill, `#E2E8F0` 1px border, r=3
  - Image with 4mm padding inside frame, centered, aspect-ratio preserved
  - If image type is not recognized (PDF uploaded): centered message card
    - "Original document uploaded as PDF" — 12pt bold `#1E293B`
    - "Open from the candidate page for PDF pages" — 8pt `#64748B`

**Overflow handling:**
- Image uses `addImageFromBytesFit` with max bounds `(ContentW-8, 217)` — already safe
- PDF fallback: text always fits in the frame

---

### Step 11 — Full Body Photo Page (Page 3)

**Functions:** `appendFullBodyPhotoSection` (line 1165), `drawFullBodyPhotoFrame` (line 1199)

**Rewrite `drawFullBodyPhotoFrame`:**
- Navy header bar at `(Margin, 12)` size `(ContentW, 14)` with "ATTACHMENT — CANDIDATE PHOTO"
- Card frame at `(Margin, 34)` size `(ContentW, 225)`:
  - `#F8FAFC` fill, `#E2E8F0` 1px border, r=3
  - Image with 4mm padding, centered, aspect-ratio preserved
  - PDF fallback: same message card pattern as passport page

**Overflow handling:**
- Same as passport page — image fitting is safe with `addImageFromBytesFit`

---

### Step 12 — Page Footer (all 3 pages)

**New function:** `drawPageFooter(pdf *gofpdf.Fpdf)`

- Thin `#E2E8F0` line at `y = PageH - 12`
- Below line, at `y = PageH - 9`:
  - Left: "Maid Recruitment Platform" — 6.5pt `#94A3B8`
  - Center: "Generated: DD/MM/YYYY" — 6.5pt `#94A3B8`
  - Right: "Page X" — 6.5pt `#94A3B8`
- Called at the end of each page after all content is drawn

**Note:** Because `auto-page-break` is enabled, the footer must be drawn manually after content (auto-page-break only adds a bottom margin, it doesn't draw content in the margin).

---

### Step 13 — Verification Checklist

**Before declaring redesign complete, verify every item:**

```
BUSINESS LOGIC (NO CHANGES)
✓ No backend endpoints modified
✓ No API calls changed
✓ No server actions modified
✓ No database schema changed
✓ No Prisma models modified
✓ No Supabase queries changed
✓ No React Query logic modified
✓ No data fetching changed
✓ No existing hooks modified
✓ No DTOs changed
✓ No response shapes modified
✓ No fields renamed
✓ No fake/duplicate/hardcoded data added
✓ No export flow changed
✓ No PDF generation pipeline changed
✓ No data parsing functions modified
✓ No formatting functions modified
✓ No image loading functions modified
✓ No domain types modified

VISUAL OUTPUT
✓ Photo renders correctly (aspect ratio preserved, centered)
✓ Photo placeholder shows when image is missing
✓ Name banner overlays photo correctly
✓ All info cards render with correct border/radius/shadow
✓ Section headers are styled consistently
✓ Icons render in correct positions
✓ Application details table renders all 9 rows
✓ Personal info 2-column grid aligns properly
✓ Passport details display correctly
✓ Work experience badges render correctly
✓ Experienced abroad table renders with #/country/duration
✓ Languages table renders with proficiency pills
✓ Branding header shows both agency names and logo
✓ Footer shows on every page

EDGE CASES
✓ Very long name (50+ chars) — shrinks, wraps, doesn't overflow
✓ Very long passport number — shrinks to fit
✓ Very long country name — shrinks in table cell
✓ Very long agency name — shrinks in header bar
✓ Very long language name — shrinks in table
✓ No skills — shows empty state message
✓ No experience abroad — shows "FIRST TIME" badge
✓ No languages — shows "FIRST TIME" badge
✓ All fields empty — all show "N/A" gracefully
✓ Photo fetch fails — shows grey placeholder
✓ Passport image is PDF — shows message card
✓ 20+ skills — grid wraps, card height adjusts
✓ 10+ languages — table extends, card height adjusts
✓ Only 1 skill/badge — no alignment issues
✓ Portrait photo — centered in frame, not stretched
✓ Landscape photo — centered in frame, not stretched

LAYOUT SAFETY
✓ No overlapping elements
✓ No clipped text
✓ No broken borders
✓ No broken icons
✓ Consistent spacing across all sections
✓ Consistent typography (size, weight, color)
✓ All cards have equal border radius
✓ All cards have equal border width
✓ All column widths add up correctly
✓ CurrentRY tracking is correct after each section
✓ Left column height matches right column (no trailing whitespace)
✓ Page 1 fits within single page for standard profiles
✓ Auto-page-break handles overflow gracefully
✓ Footer renders at correct position on every page

PDF GENERATION
✓ `go test -run TestPDFService -v` passes
✓ `go test -run TestPDFHelpers -v` passes
✓ `go build ./cmd/api/` succeeds
✓ Generated PDF opens correctly in browser
✓ Generated PDF prints at correct A4 size
✓ Generated PDF looks identical across Chrome/Firefox/Edge
```

---

## What Is NOT Modified

All data parsing, formatting, storage access, image loading, domain types, and helper functions that already work correctly remain untouched:

- `pickCandidateDocuments`, `parseLanguageEntries`, `parseExperienceEntries`, `parseSkills`, `parseJSONList`
- `formatIntPointer`, `formatListForPDF`, `manualCandidateUpper`, `manualCandidateDate`, `manualCandidateAgeValue`, `formatChildrenCount`, `shortReference`, `currentPDFDate`, `compactValue`, `candidateProfileName`, `formatExperienceLevel`
- `passportValue`, `passportDateValue`, `passportDatePointerValue`, `passportRemainingYears`, `computeRemainingYears`, `remainingYearsText`
- `decodeLogoDataURL`, `fetchRemoteAsset`, `fetchImage`, `inferContentTypeFromURL`, `normalizeImageContentType`
- `addImageFromBytes`, `addImageFromBytesFit`
- `drawFitPDFText`, `ellipsizePDFText`
- All asset cache code

**Dead code** (defined but never called): left entirely untouched.

---

## Testing After Each Step

```bash
# Unit tests
cd internal/service
go test -run TestPDFService -v
go test -run TestPDFHelpers -v

# Build verification
go build ./cmd/api/

# Full integration
# Trigger CV generation via the frontend UI
# Open generated PDF and visually inspect
```

---

## Visual Flow Diagram

```
Page 1 (Profile Sheet)
┌──────────────────────────────────────────────────┐
│  🇪🇹 Ethiopian Agency    [LOGO]    Foreign 🇸🇦   │
│  ───────────────────────────────────────────────   │
│           CANDIDATE APPLICATION PROFILE            │
│            Prepared by Ethiopian Agency            │
│                                                    │
│  ┌───LEFT COLUMN───┐  ┌───RIGHT COLUMN─────────┐  │
│  │ ┌─────────────┐ │  │ ┌─ Personal Info ─────┐ │  │
│  │ │    PHOTO    │ │  │ │ Nation  DOB  Age .. │ │  │
│  │ │   (76mm)    │ │  │ ├─ Passport ──────────┤ │  │
│  │ │  ┌───────── │ │  │ │ Pass# Iss Exp Rem  │ │  │
│  │ │  │ NAME     │ │  │ ├─ Work Experience ──┤ │  │
│  │ │  │ BANNER   │ │  │ │ ✓ Clean  ✓ Cook   │ │  │
│  │ │  └───────── │ │  │ ├─ Exp Abroad ──────┤ │  │
│  │ ├─────────────┤ │  │ │ #  Country  Dur   │ │  │
│  │ │ App Details │ │  │ ├─ Languages ───────┤ │  │
│  │ │ (9 rows)    │ │  │ │ Amh    [FLUENT]   │ │  │
│  │ └─────────────┘ │  │ └────────────────────┘ │  │
│  └─────────────────┘  └────────────────────────┘  │
│  ───────────────────────────────────────────────── │
│  Maid Recruitment  │  8/7/2026  │  Page 1         │
└──────────────────────────────────────────────────┘

Page 2 (Passport Attachment)
┌──────────────────────────────────────────────────┐
│  ATTACHMENT — PASSPORT DOCUMENT                   │
│  ┌────────────────────────────────────────────┐  │
│  │             Passport Image                 │  │
│  │         (aspect-ratio preserved)           │  │
│  └────────────────────────────────────────────┘  │
│  ───────────────────────────────────────────────── │
│  Maid Recruitment  │  8/7/2026  │  Page 2         │
└──────────────────────────────────────────────────┘

Page 3 (Photo Attachment)
┌──────────────────────────────────────────────────┐
│  ATTACHMENT — CANDIDATE PHOTO                     │
│  ┌────────────────────────────────────────────┐  │
│  │             Full Body Photo                │  │
│  │         (aspect-ratio preserved)           │  │
│  └────────────────────────────────────────────┘  │
│  ───────────────────────────────────────────────── │
│  Maid Recruitment  │  8/7/2026  │  Page 3         │
└──────────────────────────────────────────────────┘
```

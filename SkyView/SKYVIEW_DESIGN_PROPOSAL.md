# SkyView Design Proposal

## 1. Overview

SkyView is the English-only management console for the SkyBase platform. It serves as the operational UI for agent visibility, storage governance, policy management, file browsing, monitoring, and version delivery.

This proposal is based on:

- Current SkyBase backend implementation in `SkyBase/internal/handler/http/server.go`
- Current domain models in `SkyBase/internal/domain/*`
- Product scope in `docs/需求文档.md`
- Login page reference in `docs/login.vue`

The design direction is simple, calm, spacious, and operationally focused. The UI should feel reliable and organized instead of flashy.

## 2. Goals

### 2.1 Product Goals

- Present SkyBase as a control plane for distributed file collection
- Provide a clear operational dashboard for agents, storage, and policies
- Support future expansion without reworking the navigation model
- Keep the first release aligned with currently available backend capabilities

### 2.2 UX Goals

- Use low-saturation colors and quiet visual hierarchy
- Make dense operational data easy to scan
- Use consistent English labels across all pages
- Preserve a premium enterprise feel without heavy decoration

## 3. Design Language

### 3.1 Visual Keywords

- Calm
- Structured
- Industrial
- Professional
- Low-saturation
- Spacious

### 3.2 Color System

Recommended theme palette:

- Primary: `#5E7A8F`
- Primary Hover: `#4E677A`
- Primary Active: `#425867`
- Accent Deep: `#32485A`
- Success: `#5F8B78`
- Warning: `#B89863`
- Danger: `#A56A6A`
- Info: `#7A8E9E`
- Background: `#F4F7F9`
- Surface: `#FBFCFD`
- Surface Alt: `#F0F4F7`
- Border: `#D9E1E7`
- Text Primary: `#20303D`
- Text Secondary: `#6B7C88`
- Text Tertiary: `#8D9AA5`

Color usage rules:

- Use the primary color only for key actions and selected navigation
- Prefer tinted backgrounds over high-contrast solid fills
- Use semantic colors in a muted way for status tags and charts
- Keep large surfaces in neutral white and gray-blue tones

### 3.3 Typography

Recommended font stack:

```css
"Segoe UI", "Inter", "PingFang SC", "Helvetica Neue", sans-serif
```

Typography guidance:

- Page title: 28 to 32px, semibold
- Section title: 18 to 20px, semibold
- Card title: 15 to 16px, semibold
- Body text: 14px
- Secondary text: 13px
- Dense metadata: 12px

### 3.4 Shape and Depth

- Card radius: `14px`
- Input radius: `10px`
- Button radius: `10px`
- Drawer radius: `16px` on exposed edges
- Shadow style: subtle and soft

Recommended shadow:

```css
box-shadow: 0 10px 30px rgba(32, 48, 61, 0.08);
```

## 4. Login Page Direction

The login page should reference the structure and tone of `docs/login.vue`, while adapting the product messaging from VisionVault to SkyView.

### 4.1 Layout

Use a two-column full-screen layout:

- Left: brand storytelling panel
- Right: sign-in panel

Desktop ratio:

- Brand panel: about 60% to 65%
- Login panel: about 35% to 40%

Tablet and mobile:

- Collapse into a single-column layout
- Keep brand content shorter
- Move feature highlights below the login card

### 4.2 Left Brand Panel

Suggested copy:

- Eyebrow: `Distributed File Collection Control Plane`
- Product name: `SkyView`
- Main statement: `Operational visibility for agents, storage, and sync policies`
- Supporting text: `Manage SkyBase resources, monitor SkyDrop activity, and keep distributed collection workflows consistent and auditable.`

Suggested feature cards:

- `Policy Orchestration`
- `Agent Monitoring`
- `Storage Governance`

Visual guidance:

- Deep desaturated blue-gray panel
- Very soft layered gradients
- One light vertical accent rule for the message block
- Frosted feature cards with restrained contrast

### 4.3 Right Sign-in Panel

Suggested copy:

- Title: `Welcome back`
- Subtitle: `Sign in to manage SkyBase operations`

Fields:

- Username
- Password
- Verification Code
- Remember me

Footer action:

- `Download Agent Package`

### 4.4 Login Interaction

- Primary action button spans full width
- Show inline validation only when necessary
- Use subtle loading feedback
- Keep the agent package download visible but secondary

## 5. Global Application Layout

### 5.1 Shell Structure

The main application should use a classic control-plane layout:

- Left sidebar navigation
- Top application header
- Main content area
- Shared right-side drawer for details and edit forms

### 5.2 Sidebar

Width:

- Expanded: `220px`
- Collapsed: `72px`

Behavior:

- Persistent on desktop
- Overlay drawer on mobile

Visual style:

- Light neutral background
- Selected item with muted filled state
- Icon plus label structure
- Section grouping with breathing room

### 5.3 Top Header

Suggested header content:

- Product mark and system name
- Environment badge
- Optional search input
- Notification placeholder
- User menu

Header style:

- Height around `64px`
- White or near-white surface
- Bottom border rather than heavy shadow

### 5.4 Content Area

Spacing:

- Outer page padding: `24px`
- Vertical section gaps: `20px`
- Card internal padding: `20px` to `24px`

Each page should include:

- Title row
- Short description
- Page-level actions
- Main data region

## 6. Information Architecture

The first release should expose a complete operational structure even if some pages are initially static or mock-backed.

Recommended navigation:

- `Overview`
- `Agents`
- `Groups & Policies`
- `Storage`
- `Files`
- `Sync Logs`
- `Scan Reports`
- `Monitor`
- `Versions`
- `System`

### 6.1 Navigation Logic

`Overview`

- Platform summary
- Module readiness
- Quick entry to operational pages

`Agents`

- Agent inventory
- Agent details
- Heartbeat state
- Version and group assignment

`Groups & Policies`

- Group list
- Strategy definition
- Path configuration
- Filters and tag rules

`Storage`

- Storage target management
- Status and quota overview
- Local or S3 target details

`Files`

- File browsing
- Search and filtering
- Download and preview actions

`Sync Logs`

- Agent commit records
- File sync execution history

`Scan Reports`

- Directory scan results
- Tree summaries and trends

`Monitor`

- Runtime metrics
- Online and offline status
- Resource trends

`Versions`

- Current package version
- Download package records
- MD5 and rollout status

`System`

- Modules
- Platform settings
- Future RBAC entry

## 7. Page Design Details

## 7.1 Overview

Purpose:

- Provide a clean operational summary of the platform

Sections:

- Summary metric cards
- Module readiness list
- Platform relationship panel
- Recent operational timeline

Top cards:

- `System Status`
- `Enabled Modules`
- `Current Agent Version`
- `Last Agent Activity`

Data mapping:

- `/healthz`
- `/api/v1/meta/modules`

Recommended style:

- Four wide cards in a single row on desktop
- Soft monochrome icons
- Secondary metadata in muted text

## 7.2 Agents

Purpose:

- Manage the distributed SkyDrop estate

Main table fields:

- `Host Name`
- `Host SN`
- `IP Address`
- `Version`
- `Status`
- `Group`
- `Storage`
- `Last Heartbeat`
- `Last Commit`
- `Tags`

Interactions:

- Search by host, IP, or version
- Filter by status and group
- Open details drawer
- Move agent to group
- View latest heartbeat payload

Detail drawer blocks:

- Basic Information
- Runtime Status
- Storage Metrics
- Policy Snapshot
- Recent Activity

Model alignment:

- `SkyBase/internal/domain/agent/models.go`

## 7.3 Groups & Policies

Purpose:

- Represent the strategy center described in the requirements

Page structure:

- Left group list
- Right group detail tabs

Tabs:

- `Basic`
- `Storage Policy`
- `Collection Paths`
- `Filters`
- `Tag Extraction`
- `Retention`
- `Work Schedule`
- `Alerts`

Key fields:

- Name
- IP Range
- Storage ID
- Path Prefix
- Run Time
- Max Workers
- Del Time
- Work Start Time
- Work End Time
- Filter Rule
- Regex Rule
- Image Process Rule
- Alarm Group

## 7.4 Storage

Purpose:

- Manage local and S3-compatible storage targets

Recommended layout:

- Summary cards on top
- Main table below
- Detail drawer for edit and inspection

Summary cards:

- `Total Storage Targets`
- `Local Targets`
- `S3 Targets`
- `Disabled Targets`

Table fields:

- `Name`
- `Type`
- `Endpoint`
- `Bucket / Local Path`
- `Region`
- `Quota`
- `Status`
- `Updated At`

Visual behavior:

- Local storage uses path-oriented iconography
- S3 uses bucket-oriented iconography

## 7.5 Files

Purpose:

- Provide a unified browser for object and local storage

Layout:

- Top toolbar for filters and actions
- Left tree panel
- Right file table
- Breadcrumb above content table

Toolbar:

- Storage selector
- Search input
- Type filters
- Tag filters
- Action buttons

File table columns:

- `Name`
- `Path`
- `Type`
- `Size`
- `Tags`
- `Modified At`
- `Storage`

Actions:

- Preview
- Download
- Delete
- Batch Download
- Download to Server

## 7.6 Sync Logs

Purpose:

- Show agent commit records and sync execution history

Top summary:

- `Runs Today`
- `Transferred Files`
- `Transferred Size`
- `Error Count`

Table fields:

- `Agent IP`
- `Path`
- `Start Time`
- `File Count`
- `File Size`
- `Error Count`
- `Log Path`
- `Commit Time`

## 7.7 Scan Reports

Purpose:

- Display scan task records and directory inventory results

Page structure:

- Record list
- Report detail panel
- Directory tree analytics

Main filters:

- Agent
- Group
- Time range

## 7.8 Monitor

Purpose:

- Provide the main runtime operations view

Top cards:

- `Online Agents`
- `Offline Agents`
- `High CPU Alerts`
- `Storage Pressure`

Main content:

- CPU trend chart
- Memory trend chart
- Disk usage distribution
- Latest heartbeat ranking
- Version distribution

Style guidance:

- Charts should use quiet colors
- No neon or highly saturated monitor palette
- Emphasis should come from hierarchy, not brightness

## 7.9 Versions

Purpose:

- Manage SkyDrop package visibility and rollout awareness

Key modules:

- Current active version card
- Package history table
- Download package details

Primary fields:

- `Version`
- `Package ID`
- `Filename`
- `MD5`
- `Status`
- `Updated At`

Current backend alignment:

- `GET /sky/agent/version`
- `GET /sky/agent/download`

## 7.10 System

Purpose:

- Surface platform metadata and future administration entries

Suggested blocks:

- System health
- Module list
- Environment information
- Future configuration placeholders

Current backend alignment:

- `/healthz`
- `/api/v1/meta/modules`

## 8. Component Standards

### 8.1 Cards

- Title on top left
- Optional secondary action on top right
- Main metric or content below
- Quiet border and subtle background separation

### 8.2 Tables

- Medium density
- Sticky header when needed
- Row hover with low-contrast background
- Status columns use tags instead of plain text

### 8.3 Forms

- Two-column form layout on desktop for longer dialogs
- Single-column on mobile
- Group fields by business meaning
- Use helper text for sensitive fields such as secrets

### 8.4 Drawers and Modals

- Use right-side drawer for edit and inspect flows
- Use modal only for focused confirmations or short forms

### 8.5 Status Tags

Recommended visual tokens:

- Online: muted green
- Offline: cool gray
- Warning: muted amber
- Error: muted red
- Planned: slate blue-gray

## 9. English Copy Guidelines

Copy should be concise, formal, and operational.

Preferred style:

- `Agent Health`
- `Collection Policy`
- `Storage Target`
- `Recent Sync Activity`
- `No records available`

Avoid:

- Marketing slogans inside business pages
- Casual language
- Mixed Chinese and English labels

## 10. Responsive Strategy

Desktop first, but with complete tablet support.

Breakpoints:

- `>= 1440px`: wide dashboard layout
- `1200px - 1439px`: standard desktop
- `768px - 1199px`: tablet
- `< 768px`: mobile fallback

Responsive rules:

- Collapse multi-column pages progressively
- Keep tables horizontally scrollable when needed
- Convert some metric rows into stacked cards on smaller screens

## 11. Arco Design Implementation Strategy

Use Arco Design as the primary component framework.

Recommended component mapping:

- Layout shell: `a-layout`, `a-layout-sider`, `a-layout-header`, `a-layout-content`
- Navigation: `a-menu`
- Cards: `a-card`, `a-statistic`
- Tables: `a-table`
- Forms: `a-form`, `a-input`, `a-select`, `a-checkbox`, `a-time-picker`
- Detail overlays: `a-drawer`, `a-modal`
- Tags and alerts: `a-tag`, `a-alert`, `a-message`
- Tabs: `a-tabs`
- Charts: third-party charting library with customized theme

Theme strategy:

- Override Arco tokens globally
- Use CSS variables for project-level palette
- Keep component shadows and surfaces consistent across custom and Arco elements

## 12. Backend Mapping

### 12.1 Currently Available Endpoints

Based on the current SkyBase implementation:

- `GET /`
- `GET /healthz`
- `GET /api/v1/meta/modules`
- `POST /sky/agent/heartbeat`
- `GET /sky/agent/commit`
- `POST /sky/agent/scan/commit`
- `GET /sky/agent/version`
- `GET /sky/agent/download`

### 12.2 Frontend Usage in Phase 1

Pages that can directly use current backend data:

- `Overview`
- `Monitor`
- `Versions`
- `System`

Pages that should be designed now and connected later:

- `Agents`
- `Groups & Policies`
- `Storage`
- `Files`
- `Sync Logs`
- `Scan Reports`

### 12.3 Expected Response Format

SkyBase currently returns a unified JSON envelope:

```json
{
  "code": 0,
  "msg": "success",
  "data": {}
}
```

Frontend API wrappers should standardize around:

- `code === 0` as success
- `msg` as message display source
- `data` as payload body

## 13. Recommended Phase Plan

### Phase 1

- Project scaffold with Vue 3 and Arco Design
- Theme tokens and global layout
- Login page
- Overview page
- System page
- Versions page

### Phase 2

- Agents page
- Monitor page
- Static Groups and Storage page structure
- API service abstraction

### Phase 3

- Files page
- Sync Logs
- Scan Reports
- Editable management flows

## 14. Deliverables for the Next Implementation Step

The next coding step should create:

- Vue 3 app scaffold for `SkyView`
- Arco Design integration
- Global theme tokens with low-saturation palette
- Login page aligned to this proposal
- Main dashboard shell
- Initial pages for `Overview`, `Monitor`, `Versions`, and `System`

## 15. Summary

SkyView should be implemented as a composed and reliable English-language operations console. The interface should inherit the split-screen confidence of the reference login page, while the main application should use a restrained, spacious, card-based control-plane layout. This proposal keeps the visual system aligned with the current SkyBase backend and leaves enough room for the broader platform scope defined in the requirements document.

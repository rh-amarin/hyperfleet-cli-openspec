## ADDED Requirements

### Requirement: Form panel layout
The dashboard SHALL display a form panel that slides in from the right, visually consistent with the detail panel. The form panel SHALL have a fixed width of 420px and SHALL appear to the right of the detail panel when open. The form panel SHALL have a header with title "Report Adapter Status" and a close (×) button.

#### Scenario: Form panel opens on adapter click
- **WHEN** user clicks an adapter status block in the detail panel
- **THEN** the form panel slides in with width 420px and is pre-filled with that adapter's data

#### Scenario: Form panel closes
- **WHEN** user clicks the close button or the Cancel button
- **THEN** the form panel slides out (width returns to 0) and form state is cleared

#### Scenario: Form panel opens blank via + Report button
- **WHEN** user clicks the "+ Report" button in the Adapter Statuses section header
- **THEN** the form panel opens with empty fields (adapter name blank, generation 0, no conditions)

### Requirement: Adapter name field
The form SHALL include a text input for the adapter name. When pre-filled from a click, the adapter name SHALL be populated with the clicked adapter's name. The field SHALL remain editable.

#### Scenario: Pre-fill adapter name
- **WHEN** user clicks adapter block for adapter "cl-job"
- **THEN** the adapter name input is populated with "cl-job"

#### Scenario: Blank form adapter name
- **WHEN** form is opened via "+ Report"
- **THEN** the adapter name input is empty

### Requirement: Observed generation numeric stepper
The form SHALL include an observed generation field rendered as a numeric stepper with a decrement (−) button, a read-only display of the current value, and an increment (+) button. The value SHALL be a non-negative integer. The − button SHALL be disabled when the value is 0.

#### Scenario: Increment generation
- **WHEN** user clicks the + button
- **THEN** the displayed generation value increases by 1

#### Scenario: Decrement generation
- **WHEN** user clicks the − button and value is greater than 0
- **THEN** the displayed generation value decreases by 1

#### Scenario: Decrement disabled at zero
- **WHEN** the generation value is 0
- **THEN** the − button is disabled and cannot decrease the value below 0

#### Scenario: Pre-fill generation
- **WHEN** user clicks an adapter block with observed_generation 3
- **THEN** the generation stepper displays 3

### Requirement: Condition list
The form SHALL display a dynamic list of conditions. Each condition row SHALL contain: a type text input, a status radio group with options True / False / Unknown, an optional reason text input, an optional message text input, and a remove (×) button. The form SHALL include an "+ Add condition" button that appends a new empty condition row.

#### Scenario: Pre-fill conditions
- **WHEN** user clicks an adapter with two conditions
- **THEN** the form shows two condition rows populated with each condition's type, status, reason, and message

#### Scenario: Status radio buttons
- **WHEN** a condition has status "False"
- **THEN** the False radio button is selected for that condition row

#### Scenario: Add condition row
- **WHEN** user clicks "+ Add condition"
- **THEN** a new empty condition row is appended with True pre-selected as default status

#### Scenario: Remove condition row
- **WHEN** user clicks × on a condition row
- **THEN** that condition row is removed from the list

### Requirement: Form submission
Submitting the form SHALL construct an `AdapterStatusCreateRequest` JSON body with adapter name, observed_generation, observed_time set to the current UTC instant, and the conditions list. The form SHALL POST to `/api/clusters/{id}/statuses` using the currently selected cluster's ID. On 2xx response the form SHALL show a success banner and the detail panel SHALL refresh. On non-2xx the form SHALL show an error banner with the upstream error message. The Submit button SHALL be disabled while the request is in flight.

#### Scenario: Successful submission
- **WHEN** user fills the form and clicks Submit
- **THEN** a POST is sent to /api/clusters/{id}/statuses, a success banner is shown, and the detail panel refreshes

#### Scenario: Upstream error on submission
- **WHEN** the upstream API returns a 4xx or 5xx response
- **THEN** the form shows an error banner with the detail from the response body, and the form remains open

#### Scenario: Submit button disabled during flight
- **WHEN** the form submission is in progress
- **THEN** the Submit button is disabled until the response is received

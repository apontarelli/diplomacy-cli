# Diplomacy CLI — Order Format Specification

This document defines the structure and syntax for representing player orders in the Diplomacy CLI engine.

---

## 📁 File Format and Storage

### 📂 Directory Layout

- Orders are stored per-player in the path:
  ```
  {user_data_dir}/diplomacy-cli/games/{game_id}/orders/`
  ```

- Each file is named:
  ```
  {nation_id}_orders.json
  ```

Example:
```
~/.local/share/diplomacy-cli/games/classic_game_001/orders/ENG_orders.json
```

### 📄 File Contents

- Each file contains a **JSON array** of raw user input strings. 

Example:
```json
[
  "par - pic",
  "bur s par - pic",
  "mun hold"
]
```

### ♻️ Order Lifecycle

- Player submits orders to their own file.
- On resolution:
  - Orders are loaded, validated, and merged into a dictionary:
    ```json
    {
      "ENG": [...],
      "FRA": [...],
      "GER": [...]
    }
    ```
  - The merged structure is saved into the turn history:
    ```
    data/saves/{game_id}/history/{turn_code}.json
    ```
  - Individual files in `/orders/` are deleted.

---

## 🔤 Order String Syntax

### ✅ Normalization Rules

All orders are:
- **Case-insensitive**
- Trimmed of leading/trailing whitespace
- Ignored extra internal whitespace
- Ignored irrelevant punctuation

### 📌 Canonical Form

Before validation, all orders are normalized into canonical form:

```text
"  BUR  S   PAR - PIC! " → "bur s par - pic"
```

Allowed symbols:
- `-` → Movement or convoy path separator
- `/` → Coastal designation (e.g., `stp/sc`)

---

## 🧠 Order Types and Formats

### 1. Hold

A unit holds its current position.

- Format:
  ```
  {province} hold
  ```
- Example:
  ```
  mun hold
  ```

---

### 2. Move / Attack

A unit attempts to move to an adjacent province.

- Format:
  ```
  {origin} - {destination}
  ```
- Example:
  ```
  par - pic
  ```

---

### 3. Support Hold

A unit supports another unit holding in place.

- Format:
  ```
  {supporting_province} s {support_origin}
  ```
- Example:
  ```
  bur s par
  ```

---

### 4. Support Move

A unit supports a move into a province.

- Format:
  ```
  {origin} s {support_origin} - {support_destination}
  ```
- Example:
  ```
  bur s par - pic
  ```

---

### 5. Convoy

A fleet convoys an army from one province to another.

- Format:
  ```
  {origin} c {convoy_origin} - {convoy_destination}
  ```
- Example:
  ```
  bal c pru - den
  ```

---

### 6. Convoyed Move

An army attempts a move via a fleet convoy chain.

- Format:
  ```
  {origin} - {destination}
  ```
- Example:
  ```
  pru - den
  ```

---

### 7. Build

A new unit is built in a home supply center (during build phase).

- Format:
  ```
  build {army|fleet} {origin}
  ```
- Example:
  ```
  build army par
  ```

---

### 8. Disband

A unit is removed from the board (during disband phase).

- Format:
  ```
  disband {army|fleet} {origin}
  ```
- Example:
  ```
  disband fleet bal
  ```

---

## 🧪 Validation Expectations

- All orders must pass:
  - **Syntax validation**: Structural correctness
  - **Semantic validation**: Legality within current turn context

---

## ❌ Invalid Syntax Examples

- `"par -- pic"` (invalid delimiter)
- `"disband horse par"` (invalid unit type)
- `"unknowncommand"` (unrecognized order)

---

## 📝 Notes

- Province identifiers may be:
  - **Shorthand**: `par`, `bur`, `stp/sc`
  - **Long form** (optional): `paris`, `burgundy`, `st petersburg south coast`
- Coast-specific provinces (e.g., `stp/sc`) are treated as distinct
- All orders are associated with the **current turn and phase**

---

## 🧮 Grammar (EBNF)

This defines the formal grammar for interpreting order strings during validation and resolution.
All strings are case-insensitive and must be tokenized before parsing.

```ebnf
order         ::= hold
               | move
               | support_hold
               | support_move
               | convoy
               | build
               | disband

hold          ::= province "hold"
move          ::= province "-" province
support_hold  ::= province "s" province
support_move  ::= province "s" province "-" province
convoy        ::= province "c" province "-" province
build         ::= "build" unit_type province
disband       ::= "disband" unit_type province

unit_type     ::= "army" | "fleet"
province      ::= string
```

**Notes**:
- `province` refers to a valid province identifier (e.g. `par`, `bur`, `stp/sc`).
- This grammar does *not* enforce semantic rules like adjacency, phase legality, or unit ownership — only syntactic structure.
- Long-form provinces (e.g., `"burgundy"`) may be accepted at input time and resolved during validation.

---

For validator implementation and result schema, see:  
➡ `docs/VALIDATION.md` (to be defined)
```

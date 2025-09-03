# MongoDB Analyze

Analyze a MongoDB collection to discover the data types present for every field (including nested fields), how often each type occurs, and quickly identify inconsistencies across documents.

### What it does
- Scans the entire collection
- Tracks per-field data types with counts and occurrence percentages
- Supports nested fields to a configurable depth
- Outputs pretty table or JSON
- Can write results to a file

## Requirements
- Go 1.23

## Install
### Using go install
```bash
go install github.com/tomz197/mongodb-analyze/cmd/cli@latest
```
This installs the `cli` binary on your `GOBIN`/`GOPATH/bin`.

### From source
```bash
git clone https://github.com/tomz197/mongodb-analyze.git
cd mongodb-analyze
make build
```
The binary will be built at `./bin/cli`.

## Quick start
```bash
./bin/cli -collection <collection_name>
```
Specify connection details if needed:
```bash
./bin/cli -uri "mongodb://localhost:27017" -db test -collection <collection_name>
```

## Example output (table)
```text
 Name            | Type                          | Count      | Occurrence[%]
-------------------------------------------------------------------------------
 _id             | objectID                      | 9653       | 100.00
 commit          | null                          | 10         | 0.10
 commit          | string                        | 2971       | 30.78
 date_created    | UTC datetime                  | 9651       | 99.98
 date_finished   | UTC datetime                  | 8187       | 84.81
 description     | null                          | 223        | 2.31
 description     | string                        | 9192       | 95.22
 git             | embedded document             | 6502       | 67.36
  > branch       | string                        | 6500       | 67.34
  > commit       | string                        | 6502       | 67.36
  > remote_url   | string                        | 6502       | 67.36
 graph           | embedded document             | 2956       | 30.62
  > aug_graph    | binary                        | 73         | 0.76
  > aug_node_map | binary                        | 72         | 0.75
  > aug_nodes    | array[array]                  | 18         | 0.19
  > aug_nodes    | binary                        | 111        | 1.15
  > aug_nodes    | array[string]                 | 70         | 0.73
  > graph        | binary                        | 2956       | 30.62
  > nodes        | array[string]                 | 2827       | 29.29
 hostname        | string                        | 9494       | 98.35
 tag             | string                        | 139        | 1.44
 tag             | array[string]                 | 2838       | 29.40
 tag             | array[32-bit integer, string] | 4          | 0.04
 tags            | array[string]                 | 6155       | 63.76
 tags            | array[]                       | 318        | 3.29
```

Where:
- Name: Field name
  - ` > ` denotes a nested field (embedded document)
  - arrays are displayed as `array[type1, type2, ...]`
- Type: Data type of the field
- Count: Number of documents that contain this field/type
- Occurrence[%]: Percentage of documents that contain this field/type

From the above example, `tag` is inconsistent across documents: sometimes a string, sometimes an array, and there are also fields named `tags`.

## JSON output
Produce JSON instead of a table:
```bash
./bin/cli -collection <collection_name> -json
```
Small example of the JSON shape:
```json
{
  "name": "<collection>",
  "count": 9653,
  "types": [
    { "type": "objectID", "count": 9653, "name": "_id" }
  ],
  "children": [
    {
      "name": "git",
      "types": [{ "type": "embedded document", "count": 6502 }],
      "children": [
        { "name": "branch", "types": [{ "type": "string", "count": 6500 }] }
      ]
    }
  ]
}
```

## Flags

| Flag | Default | Description | Required |
|------|---------|-------------|----------|
| `-collection` | — | Name of the MongoDB collection to analyze | Yes |
| `-uri` | `mongodb://localhost:27017` | MongoDB connection URI | No |
| `-db` | `test` | MongoDB database name | No |
| `-json` | `false` | Print results in JSON instead of a table | No |
| `-depth` | `0` | Maximum depth to analyze (0 for all levels) | No |
| `-output` | `stdout` | Output file path; if set, also prints "Result saved to ..." | No |

Notes:
- `-depth 0` means no limit; the entire embedded document tree is analyzed.
- When `-output` is provided, results are written to the file. If omitted, output goes to stdout.
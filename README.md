# MongoDB Analyze

MongoDB Analyze is a tool to analyze the MongoDB collection.
It goes through entire collection and finds the data types of each field and the count of each data type.

Results are displayed in a tabular format or json if flag is provided.

Example:
```bash
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
   - ` > `: Nested field (embedded document)
   - arrays are displayed as `array[type1, type2, ...]`
- Type: Data type of the field
- Count: Number of objects that have this field
- Occurrence[%]: Percentage of how many objects have this field

From the above example, we can for example see that the field `tag` is inconsistent, it is sometimes a string and sometimes an array and was named `tags` in some cases.

## Requirements
- Go 1.23

## Usage
1. Clone the repository
2. Install the requirements
```bash
go get
```
3. Build the tool
```bash
make build
```
4. Run the tool
```bash
./bin/cli -collection <collection_name>
```

### Flags:

Required flags:
- `-collection`: Name of the collection to analyze

Optional flags:
- `-uri`: MongoDB URI, default is `mongodb://localhost:27017`
- `-db`: Name of the database to analyze, default is `test`
- `-json`: Output the results in json format
- `-depth`: Maximum depth to analyze nested fields
- `-output`: Output file name, default is `stdout`

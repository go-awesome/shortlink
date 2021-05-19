We have multiple option here based on traffic in-flow

### Single Server:

- Running combined main binary and forget everything

### Or, Separate DB server:

- We will compile and run "db" app separately which will take request from the remote/local app for read and write operation
- We can therefore execute multiple instance of app in shared basis architecture for the app "read", "write" operation.

### Or, Centralized DB server with multiple machine for API:

- Each machine has ID hard coded in the constant variable inside `constant.go` or we can also get one from `os.Getenv("MachineID")`
- Runing multiple instance of server with different Machine ID
- Running centralized db server

### Or, Each Node own DB server with multiple machine for API.

    Limitation: Each URL has analytics different on each Node as each one of them has their own DB. Useful to run on region based deployment

- Each machine has ID hard coded in the constant variable or we can also get one from 'os.Getenv("MachineID")'
- Runing multiple instance of server with different Machine ID
- Each Instance has their own DB file with unique identifier for "unique id" with value of "Machine ID"
- When "Analytics" fetch, check uniqueID - Machine ID value and then communicating with the "Fetched" machine server.
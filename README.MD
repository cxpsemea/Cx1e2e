# Repository Update
This repository will remain public but is likely to move to a new location in the future.
Further details will be provided once they are available - apologies for the inconvenience.


## Module information
This is an automated end-to-end workflow testing tool for Cx1 via the REST API, as exposed by Cx1ClientGo (https://github.com/cxpsemea/Cx1ClientGo/)

# Quick Start 
Usage:
```
    cx1e2e.exe --config tests.yaml --apikey APIKey
    cx1e2e.exe --config tests.yaml --cx1 Cx1URL --iam IAMURL --tenant Tenant --client ClientID --secret ClientSecret
```

Multiple example test.yaml definitions can be found in the examples directory. To quickly try out an example configuration, you can do the following:

``` 
    git clone https://github.com/cxpsemea/cx1e2e
    cd cx1e2e
    go run . --config examples/all.yaml --apikey <API Key> --threads 4
```

Run cx1e2e.exe -h for a list of command-line arguments.

# Test configuration
## Credentials

The IAMURL, Cx1URL, and Tenant parameters can be supplied through the command-line. This is the preferred approach when dealing with multiple Cx1 environments (eg: INT, DEV, Stage, Prod) so that the tests can be re-used easily. The proxy URL is optional and can be used for debugging.

The beginning of each test.yaml file can include standard items needed to connect to a Cx1 environment:
```
    IAMURL: https://eu.iam.checkmarx.net
    Cx1URL: https://eu.ast.checkmarx.net
    Tenant: your_tenant_here
    #ProxyURL: http://127.0.0.1:8080
```

## Test Sets

Tests are defined in Test Sets, each of which is named and can have a number of objects targeted for testing. Test Sets are executed in order, and tests within a set are executed such that all [C]reate operations are run first, then [R]ead, then [U]pdate, then [D]elete. Tests can have an optional Wait which causes the tests to pause for the specified number of seconds before continuing - this is to avoid getting blocked for spamming the API.

Refer to the existing examples in the examples folder for a list of tests and their configuration options. The examples have been split according to the object being tested (eg: groups, roles) and each test set is structure with a "create.yaml" to create various objects, "update.yaml" to update the created objects, and finally a "delete.yaml" to remove the created objects. Running the "all.yaml" test sets should clean up after itself so that there are no left-over artifacts.

### Simple test set example

A simple test to create a new role, create a new group with that role, and create a new user with the new role and also membership in the new group:
```
    Tests:
      - Name: simple test to create role & group & user
        Roles:
          - Name: e2e-test-role1
            Permissions: [ view-scans-if-in-group ]
            Test: C
        Groups:
          - Name: e2e-test-group1
            Roles: [ e2e-test-role1 ]
            Test: C
        Users:
          - Name: e2e-test-user1
            Email: e2e_test_user1@cx.local
            Roles: [ e2e-test-role1 ]
            Groups: [ e2e-test-group1 ]
            Test: C
```
In each of the above tests, the parameter "Test" includes the letter "C" - this indicates the test is to Create the relevant object.

### More complex test set example

More complicated set of 3 tests to create, read & update (set a permission), then delete a role:
```
    Tests:
      - Name: simple test Create Role
        Roles:
        - Name: e2e-test-role1
          Test: C
      - Name: simple test Read & Update Role
        Roles:
        - Name: e2e-test-role1
          Permissions: [ view-scans-if-in-group ]
          Test: U
      - Name: simple delete for all
        Roles:
        - Name: e2e-test-role1
          Test: D
```
In the tests above, each "Test" parameter includes different combinations of CRUD operations. Tests are selected through the "Test: CRUD" line which defines if the [C]reate, [R]ead, [U]pdate, or [D]elete tests are performed. 

### Test sets across multiple YAML files

Tests can also be defined in multiple files which are then referenced from the primary file. For example:
```
    Tests:
      - Name: my-special-tests
        File: special/tests.yaml
```
This will load the indicated special/tests.yaml file and add the tests to the end of the set. 

## Coverage

Currently this testing tool covers the following objects:
- Access Management (CRUD)
- Applications (CRUD)
- Groups (CRUD)
- OIDC Clients (CRUD)
- Presets (CRUD)
- Projects (CRUD)
- Queries (CRUD)
- Reports (C)
- Results (U)
- Roles (CRUD)
- Scans (CRD)
- Users (CRUD)

There are some limitations in this tool due to unimplemented functionality in the testing tool or in the underlying Cx1ClientGo library. Feel free to contribute or raise Issues.

## Engines

Scans can be triggered for: sast, sca, iac (kics), 2ms (secrets), apisec, and containers. See the examples\scan\ folder.
Custom queries & presets can be for sast and iac (kics) - however IAC custom preset & query tests require the NEW_PRESET_MANAGER_ENABLED flag currently. See the examples\sastquery and examples\iacquery folders.
Checking results can be done for: sast, sca, and iac (kics). Only SAST and KICS results can be updated. See the examples\results\ folder.

## Example output

```
[INFO][2023-03-14 16:42:18.758] Created Cx1 client my_tenant on https://deu.ast.checkmarx.net
[INFO][2023-03-14 16:42:18.892] PASS [0.133s]: Create Group Test 'simple test Create all' #1 (e2e-test-group1)
[INFO][2023-03-14 16:42:19.298] PASS [0.406s]: Create Application Test 'simple test Create all' #1 (e2e-test-app1)
[INFO][2023-03-14 16:42:19.406] PASS [0.108s]: Create Project Test 'simple test Create all' #1 (e2e-test-project1)
[INFO][2023-03-14 16:42:19.685] PASS [0.279s]: Create Role Test 'simple test Create all' #1 (e2e-test-role1)
[INFO][2023-03-14 16:42:20.755] PASS [1.070s]: Create User Test 'simple test Create all' #1 (e2e-test-user1 (e2e_test_user1@cx.local))
[ERROR][2023-03-14 16:42:20.755] FAIL [0.000s]: Create Query Test 'simple test Create all' #1 (Java -> e2e-test-querygroup1 -> e2e-test-query1) - not implemented
[INFO][2023-03-14 16:42:38.293] PASS [17.538s]: Create Preset Test 'simple test Create all' #1 (e2e-test-preset1)
[INFO][2023-03-14 16:42:38.632] Polling status of scan 7f41093f-6e2e-47b3-9d97-058abd64653e
[INFO][2023-03-14 16:42:48.747]  - Running
...
```
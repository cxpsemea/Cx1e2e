This is an automated end-to-end workflow testing tool for Cx1 via the REST API, as exposed by Cx1ClientGo (https://github.com/cxpsemea/Cx1ClientGo/)

Usage:
```
    cx1e2e.exe tests.yaml APIKey
    cx1e2e.exe tests.yaml APIKey Cx1URL IAMURL Tenant
```
Multiple example test.yaml definitions can be found in the examples directory.

The beginning of each test.yaml file will include standard items needed to connect to a Cx1 environment:
```
    IAMURL: https://eu.iam.checkmarx.net
    Cx1URL: https://eu.ast.checkmarx.net
    Tenant: your_tenant_here
    #ProxyURL: http://127.0.0.1:8080
```

The proxy URL is optional and can be used for debugging.

Tests are defined in Test Sets, each of which is named and can have a number of objects targeted for testing.

Simple test to create a new role, create a new group with that role, and create a new user with the new role and also membership in the new group:
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
          Test: RU
      - Name: simple delete for all
        Roles:
        - Name: e2e-test-role1
          Test: RD
```
In the tests above, each "Test" parameter includes different combinations of CRUD operations. Tests are selected through the "Test: CRUD" line which defines if the [C]reate, [R]ead, [U]pdate, or [D]elete tests are performed. 

Tests can also be defined in multiple files which are then referenced from the primary file. For example:
```
    Tests:
      - Name: my-special-tests
        File: special\tests.yaml
```
This will load the indicated special\tests.yaml file and add the tests to the end of the set.

Currently this testing tool covers the following objects:
- Applications (CRUD)
- Groups (CRUD)
- Presets (CRUD)
- Projects (CRUD)
- Queries (Read only)
- Results (RU)
- Roles (CRUD)
- Scans (CR)
- Users (CRUD)

There are some limitations in this tool due to unimplemented functionality in the testing tool or in the underlying Cx1ClientGo library. Feel free to contribute (PR against dev branch please) or raise Issues.

Example output:
```
[INFO][2023-03-14 16:42:18.758] Created Cx1 client cx_tam_appsec_canary_michael_kubiaczyk on https://deu.ast.checkmarx.net
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
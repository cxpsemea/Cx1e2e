This is an automated end-to-end workflow testing tool for Cx1 via the REST API, as exposed by Cx1ClientGo (https://github.com/cxpsemea/Cx1ClientGo/)

Usage:
    cx1e2e.exe tests.yaml APIKey
    cx1e2e.exe tests.yaml ClientID ClientSecret

An example of a tests.yaml file can be found in all_crud.yaml.example

The beginning of each test.yaml file will include standard items needed to connect to a Cx1 environment:
    IAMURL: https://eu.iam.checkmarx.net
    Cx1URL: https://eu.ast.checkmarx.net
    Tenant: your_tenant_here
    #ProxyURL: http://127.0.0.1:8080

The proxy URL is optional and can be used for debugging.

Tests are defined in Test Sets, each of which is named and can have a number of objects targeted for testing.

Simple test to create, read, update (but no change is specified), and delete a role:
```
    Tests:
      - Name: simple test to create role & group & user
        Roles:
          - Name: e2e-test-role1
            Permissions: [ view-scans-if-in-group ]
            Test: CRUD
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

Tests are selected through the "Test: CRUD" line for each test which decides if the [C]reate, [R]ead, [U]pdate, or [D]elete tests are performed. 

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
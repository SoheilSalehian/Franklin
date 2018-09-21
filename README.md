# Franklin 

[![Build Status](https://travis-ci.com/SoheilSalehian/Franklin.svg?branch=master)](https://travis-ci.com/SoheilSalehian/Franklin) 
[![Coverage Status](https://coveralls.io/repos/github/SoheilSalehian/Franklin/badge.svg?branch=master)](https://coveralls.io/github/SoheilSalehian/Franklin?branch=master)

A simple REST API server that will expose endpoints to allow users to place and modify orders.

## User Stories 

- As a user I would like be able to signup with my username, pass and zipcode 
- As a user I would like to signin.
- As a user I would like submit orders of items.
- As a user I would like update my order.
- As a user I would like remove my order.
- As a user I would like list all my orders.
- As a user I would like to know the closest walmart store to my zipcode.

- As a admin user I would like to get user information of any user.

- As a service user passwords should be encrypted in the database.
- As a service users can only interact with their own orders.
- As a service all request needs to be authenticated.

## API Specs

[API Docs](https://documenter.getpostman.com/view/5413928/RWaPs5t6#bfd698f3-1837-4a0d-8ce7-49f68252f1da)


## Technical Stack

We are using a Golang based API layer, interacting with a sqlite3 database.


## Quick Setup

1. Make sure Go 1.11+ and Go's dependancy management tool (dep) is installed.

2. Get all the dependancies by:
```
dep ensure
```

3. Set environment variable for Walmart's open API key:
```
export WALMART_OPEN_API_KEY= ...
```

4. Build:
```
go build
```

5. Run the web server:
```
./Franklin
```

6. Ping endpoints 

## Tests
1. Make sure Go 1.11+ and Go's dependancy management tool (dep) is installed.

2. Get all the dependancies by:
```
dep ensure
```
3. Run tests:
```
go test
```

## Libraries Used
No frameworks and minimal libraries used for fun and profits :)

Here are my reasoning for the important Golang frameworks used:

  ```
  name = "github.com/gorilla/mux"
  version = "1.6.2"
  ```
- Used to avoid basic http boilerplate, and re-inventing the wheel this is a must for any serious REST API project in Go.
<br><br>

```
  branch = "master"
  name = "github.com/jarcoal/httpmock"
```
- Used to mock the external API calls to Walmart for testing
  branch = "master"
  name = "github.com/jarcoal/httpmock"
<br><br>

```
  name = "github.com/mattn/go-sqlite3"
  version = "1.9.0"
```
- Basic building blocks to interact with sqlite without using any ORMs
<br><br>

```
  branch = "master"
  name = "github.com/prometheus/common"
```
  - Better logging than std library with better tracing.
<br><br>

```
  name = "github.com/stretchr/testify"
  version = "1.2.2"
```
 - Used for assertions in tests, espcially to compare json response expectations.
<br><br>

```
  branch = "master"
  name = "golang.org/x/crypto"
```
- Used to hash the passwords and compare for basicAuth







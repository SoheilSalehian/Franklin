# Franklin
A simple REST API server that will expose endpoints to allow users to place and modify orders.

## Current Usecases

The usecases currently supported are:

- Creating a new user
- Updating a user
- Deleting a user

## API Specs
resource "user":

- Create a new user in response to a valid POST request at /user
- Update a user in response to a valid PUT request at /user/{id}
- Fetch a user in response to a valid GET request at /user/{id}
- Delete a user in response to a valid DELETE request at /user/{id}

## Technical Stack
We are using a Golang based API layer, interacting with a sqlite3 database.

## Quick Setup
TBD





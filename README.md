# Puppet Summary

This repo has taken inspiration from the [puppet-summary repo](https://github.com/skx/puppet-summary). When using the
summary application I found that you could not run the application in High Availability mode within Kubernetes. This is
due to the fact that the application uses a SQLite database and the file is locked when the application is
using `SQLITE` and storing the YAML files on the local filesystem.

This repo has been updated to use MySQL, MongoDB or SQLite as the database backend and also has the ability to upload
the raw reports to Google Cloud Storage for further processing. This also allows for more than one instance of the
application to be running at the same time. By allowing for MySQL or MongoDB as the database backend, this allows for
data retention on a more reliable database.

## Usage

There are multiple ways to run the application. You can either run the application directly on your host or you can
run the application within a Docker container. The application will listen on port `8080` by default.

When running teh application, the app will default to using the local filesystem to store the raw reports, the default
database is SQLite. These can be changed by using the flags that are available (Listed below).

### Commands

The application has the following commands:

#### Serve

The `serve` command will start the application and listen on port `8080`. This is the primary command that you will
use to run the application.

You can view the help for this command by running:

```shell
./puppet-summary serve --help
```

#### Purge

The `purge` command will purge the database of data older than the specified number of days. This is useful if you
want to keep the database size down.

You can view the help for this command by running:

```shell
./puppet-summary purge --help
```

#### Version

The `version` command will print the version of the application.

This command does not have any flags and can be run by running:

```shell
./puppet-summary version
```

## Setup

#### MySQL

When using MySQL, you will be required to specify a `DB_CONN_STR` environment variable with the connection string
to your MySQL database. For example:

```text
DB_CONN_STR="root:Password01@tcp(localhost:3306)/puppet-summary?timeout=90s&multiStatements=true&parseTime=true"
```

#### MongoDB

When using MongoDB, you will be required to specify a `DB_CONN_STR` environment variable with the connection URI to your
MongoDB database. For example:

```text
DB_CONN_STR="mongodb+srv://user:password@host/?retryWrites=true"
```

#### Google Cloud Storage

```shell
./puppet-summary -gcs <bucket>
```

This will allow the `/upload` endpoint to push the raw reports from Puppet to Google Cloud Storage. For this, you will
be required to specify a `gcs-bucket` flag with the name of the bucket to upload to; you will also need
the `GCS_CREDENTIALS` environment variable with the contents of the JSON credentials file. For example:

```text
GCS_CREDENTIALS=<json-contents>
```

#### S3 Storage

_Coming soon_

#### Secure Upload

```shell
./puppet-summary -upload-token <token>
```

This will enable the `/upload` endpoint to only accept requests that have this Bearer token in the `Authorization`
header. This is useful if you want to prevent any unauthorised requests to the `/upload` endpoint. One architecture
pattern that you could use is to have a proxy in front of the Puppet Summary application that will handle the
authentication and then forward the request to the Puppet Summary application with the Bearer token in the
`Authorization` header.

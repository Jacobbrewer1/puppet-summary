# Puppet Summary

This is an application that helps you sort through the Puppet reports that are generated by your Puppet infrastructure.
The summary will give you a quick overview of the state of your infrastructure and will help you identify any issues
that may have occurred. There is an API and web interface that you can use to view the reports. The application will
also allow you to upload the raw reports from Puppet to the application, which will then be processed and stored in
the database. The application will also allow you to purge the database of old reports. The application is written in
GO and uses the gorilla/mux router for the API and the web interface.

## Usage

There are multiple ways to run the application. You can either run the application directly on your host or you can
run the application within a Docker container. The application will listen on port `8080` by default.

When running the application, the app will default to using the local filesystem to store the raw reports, the default
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

#### Endpoint Authentication

```shell
./puppet-summary -auth-token <token>
```

This will enable the security on the endpoints that use the authentication method `AuthOptionRequired`. This includes
the `/upload`. If the token is not provided, there will be no security on the endpoints.

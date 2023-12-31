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

### Flags

```text
Usage of ./puppet-summary:
  -db <string>
        Database to use (default "sqlite"). Valid options are: sqlite, mysql, mongo.
  -gcs
        Enable Google Cloud Storage upload.
  -upload-token <string>
        Enables secure upload. Requires a token to be specified. This is useful if you want to prevent any unauthorised requests to the /upload endpoint.
  -version
        Print version and exit.
```

### Setup

#### SQLite

```shell
./puppet-summary -db sqlite
```

This will create a `puppet-summary.db` file in the current directory.

#### MySQL

```shell
./puppet-summary -db mysql
```

For this, you will be required to specify a `MYSQL_CONNECTION` environment variable with the connection string to your
MySQL database. For example:

```text
MYSQL_CONNECTION="root:Password01@tcp(localhost:3306)/puppet-summary?timeout=90s&multiStatements=true&parseTime=true"
```

#### MongoDB

```shell
./puppet-summary -db mongo
```

For this, you will be required to specify a `MONGO_URI` environment variable with the connection URI to your MongoDB
database. For example:

```text
mongodb+srv://user:password@host/?retryWrites=true
```

#### Google Cloud Storage

```shell
./puppet-summary -gcs
```

This will enable the `/upload` endpoint to push the raw reports from Puppet to Google Cloud Storage. For this, you will
be required to specify a `GCS_BUCKET` environment variable with the name of the bucket to upload to; you will also need
the `GCS_CREDENTIALS` environment variable with the contents of the JSON credentials file. For example:

```text
GCS_BUCKET="puppet-reports"
```

#### Secure Upload

```shell
./puppet-summary -upload-token <token>
```

This will enable the `/upload` endpoint to only accept requests that have this Bearer token in the `Authorization`
header. This is useful if you want to prevent any unauthorised requests to the `/upload` endpoint. One architecture
pattern that you could use is to have a proxy in front of the Puppet Summary application that will handle the
authentication and then forward the request to the Puppet Summary application with the Bearer token in the
`Authorization` header.

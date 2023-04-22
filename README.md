# docker-mariadb-temporal-tables-tutorial

This is a tiny REST API to show how we can use temporal tables from [MariaDB](https://mariadb.org/).

It is ready to be deployed in [Kubernetes][https://kubernetes.io/]. It uses [Minikube](https://minikube.sigs.k8s.io/docs/) to set up a local cluster.

Make sure you check `Postman` folder; I've exported a collection with invocation examples.

## running it

```
make run
```

Then,

```
make kong
```

and use the returned host to access the app.

## running unit tests

```
make test
```

## running integration tests

```
make int-test
```

## running linter

```
make lint
```

## generating api's documentation via Swagger

```
make swagger
```

## launching swagger ui

```
make swagger
```

Then head to `localhost`.

## Kubernetes dashboard

```
make minikube-dashboard
```

## references
- [go-swagger](https://github.com/go-swagger/go-swagger)
- [MariaDB](https://mariadb.org/)
- [System-versioned tables](https://mariadb.com/kb/en/system-versioned-tables/)

## related articles
- [Golang: a RESTful api using temporal tables with MariaDB](https://www.linkedin.com/pulse/golang-restful-api-using-temporal-table-mariadb-tiago-melo/?lipi=urn%3Ali%3Apage%3Ad_flagship3_profile_view_base_post_details%3BzM9pMHNHTbe875AXAzROcA%3D%3D)
- [running a dockerized linter](https://www.linkedin.com/pulse/golang-running-dockerized-linter-tiago-melo/)
- [declarative validations](https://www.linkedin.com/pulse/golang-declarative-validation-made-similar-ruby-rails-tiago-melo/)
- [database migrations](https://www.linkedin.com/pulse/go-database-migrations-made-easy-example-using-mysql-tiago-melo/)

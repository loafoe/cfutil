CFUTIL
======
This package provides a number of convenience functions for apps running in Cloudfoundry
In case the app is started outside of Cloudfoundry an attempt is made to simulate the
Cloudfoundry environment

Simulating Cloudfoundry Services
=================================
When running locally the app looks for the following environment variables

* CF_LOCAL_POSTGRES
* CF_LOCAL_SMTP
* CF_LOCAL_RABBITMQ

Services are setup using the variable values as the URI. This allows you to use local Postgres, SMTP and RabbitMQ services just as you would in an actual Cloudfoundry deployment

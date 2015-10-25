CFUTIL
======
This package provides a number of convenience functions for apps running in Cloud Foundry
In case the app is started outside of Cloud Foundry an attempt is made to simulate the
Cloud Foundry environment

Simulating Cloud Foundry Services
=================================
When running locally the app looks for the following environment variables

* CF_LOCAL_POSTGRES
* CF_LOCAL_SMTP

The value in these variables are used to populate the uri field in the credentials section

# Assure

Assure is a system testing framework aimed at improving software testability. It provides a rich and flexible framework that is used to implement system level tests for any system.

Assure does not assume anything about the system under test (SUT) and as such does not make any assumptions about technologies or protocols. The user is responsible for implementing modules to support testing the SUT, which are then compiled together with the Assure framework to produce an executable that is used for testing. Assure does, however, have a strict view on how a system test should be monitored, see the [monitoring](#monitoring) section for more information.

## 

## Monitoring {#monitoring}

Assure approaches system monitoring in a black box manner. The following monitoring methods are available with Assure's built in monitoring tools:

 1. CPU
 2. Memory
 3. (Prometheus) metrics
 4. Structured (JSON) logs

More monitoring tools may be added in the future, but it is not something that we do on a whim. We believe that a well designed system should be perfectly possible to monitor using the above methods. Adding any sort of white box monitoring is out of the question and because of this it is not possible to implement custom modules for monitoring. We believe such tests should be done on a lower level, not on and integration/system level.

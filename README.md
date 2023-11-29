# Dependencies
## What

dependencies is a small command client tool to get (recursive) dependencies out of our applications.

## How

This tool is able to collect dependency information in 1 way.

  1. provide sisu directory

  If you provide the **-sisu** parameter, it will read the **.baur.toml** file to get the directories where to find the applications.
  Inside of the application directory it will look for a **.deps.toml** file where the dependencies of tha application is defined.

### deps.toml

Structure:
```
name = "a-service"
talks_to = [ "consul","postgres" ]
```

### What

  1. Give me an ordered list of services I need to deploy in order to do integration tests
     1.1  This list can be a comma separated list of services like this: **...  -service claim-service,pdfrender-service**


```
% ./dependencies deploy-order ~/sandbox/git/work/sisu stg eu -service actionrequest-service
deployment order for actionrequest-service service(s)
"consul"
"rabbitmq"
...
...
"actionrequest-service"
```


  2. Generate a visualization:

```
./dependencies deploy-order --format dot ~/sandbox/git/work/sisu stg eu -service certificate-service > output.dot
dot -Tsvg -o output.svg output.dot
```

# Using Lambda with local embedding database


## Import

Here some testdocuments in `import/testdata` are imported in the local embedding database and then copied to the lambda directory.

You need acces to a AWS account with Bedrock models "titan" to fetch the embeddings.

1) Change into the `import` directory.
  ```bash
  cd import
  ```
2) Run the import script.
  ```bash
  task import
  ```
3) Copy the local database file to the lambda directory.
  ```bash
  task copy
  ```


## Backend - Lambda

With AWS SAM a lambda function is created that uses the local embedding database to find similar documents.

1) Change into the `backend` directory.
  ```bash
  cd backend
  ```

2) Build the lambda function.
  ```bash
  task build
  ```

3) Deploy the lambda function.
  ```bash
  task deploy
  ```

## CLI - Test the Lambda function

This is a simple CLI to test the lambda function.
The Name of the Lambda function `hugoembedding` is hardcoded.

1) Change into the `cli` directory.
  ```bash
  cd cli
  ```
2) Build the cli.
  ```bash
  task build
  ```

3) Run the cli.
  ```bash
  ./dist/query --question "How do you start a CDK Project?"
  ```

4) See documents also
  ```bash
  ./dist/query --question "How do you start a CDK Project?" --verbose
  ```

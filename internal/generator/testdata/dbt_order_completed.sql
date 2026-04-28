version: 2

models:
  - name: Order Completed
    columns:
      - name: total
        tests:
          - dbt_expectations.expect_column_to_exist
          - not_null
      - name: userId
        tests:
          - dbt_expectations.expect_column_to_exist
          - not_null

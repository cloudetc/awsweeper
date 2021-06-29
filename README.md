<p align="center">
  <img alt="AWSweeper" src="https://github.com/jckuester/awsweeper/blob/master/img/logo.png" height="150" />
  <h3 align="center">AWSweeper</h3>
  <p align="center">A tool for cleaning your AWS account</p>
</p>

---
[![Release](https://img.shields.io/github/release/jckuester/awsweeper.svg?style=for-the-badge)](https://github.com/jckuester/awsweeper/releases/latest)
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=for-the-badge)](/LICENSE.md)
[![Travis](https://img.shields.io/travis/jckuester/awsweeper/master.svg?style=for-the-badge)](https://travis-ci.com/jckuester/awsweeper)

AWSweeper is able to clean out [over 290 resource types](#supported-resources) in your AWS account. Resources to be
deleted can be filtered by their type, ID, tags, or creation date
using [regular expressions](https://golang.org/pkg/regexp/syntax/)
declared in a YAML file (see [filter.yml](filter.yml) as an example).

To keep up supporting the continuously growing number of new resources, AWSweeper is standing upon the shoulders of
delete routines provided by the [Terraform AWS provider](https://github.com/terraform-providers/terraform-provider-aws).
List operations are borrowed from the [awsls](https://github.com/jckuester/awsls) open-source project and are
code-generated based on the [model of the AWS API](https://github.com/aws/aws-sdk-go-v2/tree/master/models/apis).

Not being fully there yet, but the goal is to support every AWS resource that is covered by Terraform
(currently over 500) without adding or maintaining much code here.

If you run into issues deleting resources, please open an issue or ping me on [Twitter](https://twitter.com/jckuester).

Happy erasing!

**New**: [`awsrm`](https://github.com/jckuester/awsrm) got released! It's a Unix-like `rm` to simplify
deleting AWS resources across multiple accounts + regions and integrates well with other Unix-like tools, such
as `grep` for filtering resources by tags or other attributes. Go and check it out!

## Example

[![AWSweeper tutorial](img/asciinema-tutorial.gif)](https://asciinema.org/a/149097)

## Features

* Nothing will be deleted without your confirmation. AWSweeper always lists all resources first and then waits for
  your approval (also without the `--dry-run` flag). With the `--dry-run` flag, AWSweeper lists all resources and exits.
* Using the `-force` flag (dangerous!), AWSweeper can in run an automated fashion without human interaction and approval,
  for example, as part of a CI pipeline

## Installation

### Binary Releases

You can download a specific version of AWSweeper on the [releases page](https://github.com/jckuester/awsweeper/releases)
or install it the following way to `./bin/`:

```bash
curl -sSfL https://raw.githubusercontent.com/jckuester/awsweeper/master/install.sh | sh -s v0.11.1
```

### Homebrew

Homebrew users can install by:

```bash
brew install jckuester/tap/awsweeper
```

For more information on Homebrew taps please see the [tap documentation](https://docs.brew.sh/Taps).

## Usage

    awsweeper [options] <filter.yml>

To see options available run `awsweeper --help`.

## Filter

Resources are deleted via a filter declared in a YAML file.

    aws_instance:
      # instance filter part 1
      - id: ^foo.*
        created:
          before: 2018-10-14
          after: 2018-06-28 12:28:39

      # instance filter part 2
      - tags:
          foo: bar
          NOT(owner): .*

    aws_security_groups:

The filter snippet above deletes all EC2 instances that ID matches `^foo.*` and that have been created between
 `2018-06-28 12:28:39` and `2018-10-14` UTC (instance filter part 1); additionally, EC2 instances having a tag
 `foo: bar` *AND* not a tag key `owner` with any value are deleted (instance filter part 2); last but not least,
 ALL security groups are deleted by this filter.

The general filter syntax is as follows:

    <resource type>:
      - id: <regex to filter by id> | NOT(<regex to filter by id>)
        tagged: bool (optional)
        tags:
          <key> | NOT(key): <regex to filter value> | NOT(<regex to filter value>)
          ...
        created:
          before: <timestamp> (optional)
          after: <timestamp> (optional)
      # OR
      - ...
    <resource type>:
      ...

Here is a more detailed description of the various ways to filter resources:

##### 1) Delete all resources of a particular type

   [Terraform resource type identifiers](https://www.terraform.io/docs/providers/aws/index.html) are used to delete
   resources by type. The following filter snippet deletes *ALL* security groups, IAM roles, and EC2 instances:

    aws_security_group:
    aws_iam_role:
    aws_instance:

   Don't forget the `:` at the end of each line.

##### 2) Delete by tags

   If most of your resources have tags, this is probably the best way to filter them
   for deletion. **Be aware**: Not all resources [support tags](#supported-resources) yet and can be filtered this way.

   The key and the value part of the tag filter can be negated by a surrounding `NOT(...)`. This allows for removing of
   all resources not matching some tag key or value. In the example below, all EC2 instances without the `owner: me`
   tag are deleted:

    aws_instance:
      - tags:
          NOT(Owner): me

   The flag `tagged: false` deletes all resources that have no tags. Contrary, resources with any tags can be deleted
   with `tagged: true`:

    aws_instance:
      - tagged: true

##### 3) Delete By ID

   You can filter resources of a particular type based on their IDs.

   To see what the IDs for a type of resource look like (sometimes it's the name, sometimes the ARN, ...), run AWSweeper
   first in dry-run mode. Then, use this information to create the YAML filter accordingly.

   The id filter can be negated by surrounding the regex with `NOT(...)`

##### 4) By creation date

   You can select resources by filtering on the date they have been created using an absolute or relative date.

   The supported formats are:
   * Relative
     * Nanosecond: `1ns`
     * Microsecond: `1us`
     * Millisecond: `1ms`
     * Second: `1s`
     * Minute: `1m`
     * Hour: `1h`
     * Day: `1d`
     * Week: `1w`
     * Month: `1M`
     * Year: `1y`
   * Absolute:
     * RCF3339Nano, short dates: `2006-1-2T15:4:5.999999999Z07:00`
     * RFC3339Nano, short date, lower-case "t": `2006-1-2t15:4:5.999999999Z07:00`
     * Space separated, no time zone: `2006-1-2 15:4:5.999999999`
     * Date only: `2006-1-2`

## Supported resources

The list below shows the 297 supported (Terraform) [resource types](https://www.terraform.io/docs/providers/aws/index.html),
which have to be used in the YAML file to filter resources by their type.

| Service / Resource Type | Delete by tag | Delete by creation date
| :-----------------------------   |:-------------:|:-----------------------:
| **accessanalyzer** |
| aws_accessanalyzer_analyzer |  x  |  |
| **acm** |
| aws_acm_certificate |  x  |  |
| **apigateway** |
| aws_api_gateway_api_key |  x  |  |
| aws_api_gateway_client_certificate |  x  |  |
| aws_api_gateway_domain_name |  x  |  |
| aws_api_gateway_rest_api |  x  |  |
| aws_api_gateway_usage_plan |  x  |  |
| aws_api_gateway_vpc_link |  x  |  |
| **apigatewayv2** |
| aws_apigatewayv2_api |  x  |  |
| aws_apigatewayv2_domain_name |  x  |  |
| aws_apigatewayv2_vpc_link |  x  |  |
| **appmesh** |
| aws_appmesh_mesh |  x  |  |
| **appsync** |
| aws_appsync_graphql_api |  x  |  |
| **athena** |
| aws_athena_named_query |  |  |
| aws_athena_workgroup |  x  |  x  |
| **autoscaling** |
| aws_autoscaling_group |  x  |  x  |
| aws_launch_configuration |  |  x  |
| **backup** |
| aws_backup_plan |  x  |  x  |
| aws_backup_vault |  x  |  x  |
| **batch** |
| aws_batch_compute_environment |  x  |  |
| aws_batch_job_definition |  x  |  |
| **cloudformation** |
| aws_cloudformation_stack |  x  |  x  |
| aws_cloudformation_stack_set |  x  |  |
| aws_cloudformation_type |  |  |
| **cloudhsmv2** |
| aws_cloudhsm_v2_cluster |  x  |  |
| **cloudtrail** |
| aws_cloudtrail |  x  |  |
| **cloudwatch** |
| aws_cloudwatch_dashboard |  |  |
| **cloudwatchevents** |
| aws_cloudwatch_event_archive |  |  x  |
| aws_cloudwatch_event_bus |  x  |  |
| **cloudwatchlogs** |
| aws_cloudwatch_log_destination |  |  x  |
| aws_cloudwatch_log_group |  x  |  x  |
| aws_cloudwatch_log_resource_policy |  |  |
| aws_cloudwatch_query_definition |  |  |
| **codeartifact** |
| aws_codeartifact_domain |  x  |  x  |
| aws_codeartifact_repository |  x  |  |
| **codebuild** |
| aws_codebuild_project |  x  |  |
| aws_codebuild_report_group |  x  |  |
| aws_codebuild_source_credential |  |  |
| **codecommit** |
| aws_codecommit_repository |  x  |  |
| **codedeploy** |
| aws_codedeploy_deployment_config |  |  |
| **codepipeline** |
| aws_codepipeline_webhook |  x  |  |
| **codestarconnections** |
| aws_codestarconnections_connection |  x  |  |
| aws_codestarconnections_host |  |  |
| **codestarnotifications** |
| aws_codestarnotifications_notification_rule |  x  |  |
| **configservice** |
| aws_config_config_rule |  x  |  |
| aws_config_configuration_aggregator |  x  |  x  |
| aws_config_configuration_recorder |  |  |
| aws_config_conformance_pack |  |  |
| aws_config_delivery_channel |  |  |
| **costandusagereportservice** |
| aws_cur_report_definition |  |  |
| **databasemigrationservice** |
| aws_dms_certificate |  x  |  |
| aws_dms_endpoint |  x  |  |
| aws_dms_replication_subnet_group |  x  |  |
| aws_dms_replication_task |  x  |  |
| **datasync** |
| aws_datasync_agent |  x  |  |
| aws_datasync_task |  x  |  |
| **dax** |
| aws_dax_parameter_group |  |  |
| aws_dax_subnet_group |  |  |
| **devicefarm** |
| aws_devicefarm_project |  |  |
| **directconnect** |
| aws_dx_connection |  x  |  |
| aws_dx_hosted_private_virtual_interface |  |  |
| aws_dx_hosted_public_virtual_interface |  |  |
| aws_dx_hosted_transit_virtual_interface |  |  |
| aws_dx_lag |  x  |  |
| aws_dx_private_virtual_interface |  x  |  |
| aws_dx_public_virtual_interface |  x  |  |
| aws_dx_transit_virtual_interface |  x  |  |
| **dlm** |
| aws_dlm_lifecycle_policy |  x  |  |
| **dynamodb** |
| aws_dynamodb_global_table |  |  |
| aws_dynamodb_table |  x  |  |
| **ec2** |
| aws_ami |  x  |  x  |
| aws_ebs_snapshot |  x  |  x  |
| aws_ebs_volume |  x  |  x  |
| aws_ec2_capacity_reservation |  x  |  x  |
| aws_ec2_carrier_gateway |  x  |  |
| aws_ec2_client_vpn_endpoint |  x  |  x  |
| aws_ec2_fleet |  x  |  x  |
| aws_ec2_local_gateway_route_table_vpc_association |  x  |  |
| aws_ec2_managed_prefix_list |  x  |  |
| aws_ec2_traffic_mirror_filter |  x  |  |
| aws_ec2_traffic_mirror_session |  x  |  |
| aws_ec2_traffic_mirror_target |  x  |  |
| aws_ec2_transit_gateway |  x  |  x  |
| aws_ec2_transit_gateway_peering_attachment |  x  |  x  |
| aws_ec2_transit_gateway_route_table |  x  |  x  |
| aws_ec2_transit_gateway_vpc_attachment |  x  |  x  |
| aws_egress_only_internet_gateway |  x  |  |
| aws_eip |  x  |  |
| aws_instance |  x  |  x  |
| aws_internet_gateway |  x  |  |
| aws_key_pair |  x  |  |
| aws_launch_template |  x  |  x  |
| aws_nat_gateway |  x  |  x  |
| aws_network_acl |  x  |  |
| aws_network_interface |  x  |  |
| aws_placement_group |  x  |  |
| aws_route_table |  x  |  |
| aws_security_group |  x  |  |
| aws_spot_fleet_request |  x  |  x  |
| aws_spot_instance_request |  x  |  x  |
| aws_subnet |  x  |  |
| aws_vpc |  x  |  |
| aws_vpc_endpoint |  x  |  x  |
| aws_vpc_endpoint_connection_notification |  |  |
| aws_vpc_endpoint_service |  x  |  |
| aws_vpc_peering_connection |  x  |  |
| aws_vpn_gateway |  x  |  |
| **ecr** |
| aws_ecr_repository |  x  |  |
| **ecrpublic** |
| aws_ecrpublic_repository |  |  |
| **ecs** |
| aws_ecs_cluster |  x  |  |
| aws_ecs_task_definition |  x  |  |
| **efs** |
| aws_efs_access_point |  x  |  |
| aws_efs_file_system |  x  |  x  |
| **eks** |
| aws_eks_cluster |  x  |  |
| **elasticache** |
| aws_elasticache_global_replication_group |  |  |
| aws_elasticache_replication_group |  x  |  |
| **elasticbeanstalk** |
| aws_elastic_beanstalk_application |  x  |  |
| aws_elastic_beanstalk_application_version |  x  |  |
| aws_elastic_beanstalk_environment |  x  |  |
| **elastictranscoder** |
| aws_elastictranscoder_pipeline |  |  |
| aws_elastictranscoder_preset |  |  |
| **elb** |
| aws_elb |  x  |  x  |
| **elbv2** |
| aws_alb_target_group |  x  |  |
| aws_lb |  x  |  x  |
| aws_lb_target_group |  x  |  |
| **emr** |
| aws_emr_security_configuration |  |  |
| **firehose** |
| aws_kinesis_firehose_delivery_stream |  x  |  |
| **fms** |
| aws_fms_policy |  |  |
| **fsx** |
| aws_fsx_lustre_file_system |  x  |  x  |
| aws_fsx_windows_file_system |  x  |  x  |
| **gamelift** |
| aws_gamelift_alias |  x  |  x  |
| aws_gamelift_build |  x  |  x  |
| aws_gamelift_fleet |  x  |  |
| aws_gamelift_game_session_queue |  x  |  |
| **globalaccelerator** |
| aws_globalaccelerator_accelerator |  x  |  x  |
| **glue** |
| aws_glue_crawler |  x  |  x  |
| aws_glue_dev_endpoint |  x  |  |
| aws_glue_job |  x  |  |
| aws_glue_ml_transform |  x  |  |
| aws_glue_registry |  x  |  x  |
| aws_glue_schema |  x  |  x  |
| aws_glue_security_configuration |  |  |
| aws_glue_trigger |  x  |  |
| aws_glue_workflow |  x  |  |
| **guardduty** |
| aws_guardduty_detector |  x  |  |
| **iam** |
| aws_iam_access_key |  |  x  |
| aws_iam_account_alias |  |  |
| aws_iam_group |  |  x  |
| aws_iam_instance_profile |  x  |  x  |
| aws_iam_policy |  x  |  x  |
| aws_iam_role |  x  |  x  |
| aws_iam_server_certificate |  x  |  |
| aws_iam_service_linked_role |  |  x  |
| aws_iam_user |  x  |  x  |
| **imagebuilder** |
| aws_imagebuilder_component |  x  |  |
| aws_imagebuilder_distribution_configuration |  x  |  |
| aws_imagebuilder_image |  x  |  |
| aws_imagebuilder_image_pipeline |  x  |  |
| aws_imagebuilder_image_recipe |  x  |  |
| aws_imagebuilder_infrastructure_configuration |  x  |  |
| **iot** |
| aws_iot_certificate |  |  x  |
| aws_iot_policy |  |  |
| aws_iot_role_alias |  |  |
| aws_iot_thing |  |  |
| aws_iot_thing_type |  |  |
| aws_iot_topic_rule |  x  |  |
| **kafka** |
| aws_msk_cluster |  x  |  x  |
| aws_msk_configuration |  |  x  |
| **kinesis** |
| aws_kinesis_stream |  x  |  |
| **kinesisanalytics** |
| aws_kinesis_analytics_application |  x  |  |
| **kinesisanalyticsv2** |
| aws_kinesisanalyticsv2_application |  x  |  |
| **kms** |
| aws_kms_external_key |  x  |  |
| aws_kms_key |  x  |  |
| **lambda** |
| aws_lambda_code_signing_config |  |  |
| aws_lambda_event_source_mapping |  |  |
| aws_lambda_function |  x  |  |
| **lexmodelbuildingservice** |
| aws_lex_bot |  |  |
| aws_lex_intent |  |  |
| aws_lex_slot_type |  |  |
| **licensemanager** |
| aws_licensemanager_license_configuration |  x  |  |
| **lightsail** |
| aws_lightsail_domain |  |  |
| aws_lightsail_instance |  x  |  |
| aws_lightsail_key_pair |  |  |
| aws_lightsail_static_ip |  |  |
| **macie2** |
| aws_macie2_classification_job |  x  |  |
| aws_macie2_custom_data_identifier |  x  |  |
| aws_macie2_findings_filter |  x  |  |
| **mediaconvert** |
| aws_media_convert_queue |  x  |  |
| **mediapackage** |
| aws_media_package_channel |  x  |  |
| **mediastore** |
| aws_media_store_container |  x  |  x  |
| **mq** |
| aws_mq_broker |  x  |  |
| aws_mq_configuration |  x  |  |
| **mwaa** |
| aws_mwaa_environment |  x  |  |
| **neptune** |
| aws_neptune_event_subscription |  x  |  |
| **networkfirewall** |
| aws_networkfirewall_firewall |  x  |  |
| aws_networkfirewall_firewall_policy |  x  |  |
| aws_networkfirewall_rule_group |  x  |  |
| **opsworks** |
| aws_opsworks_stack |  x  |  |
| aws_opsworks_user_profile |  |  |
| **qldb** |
| aws_qldb_ledger |  x  |  |
| **rds** |
| aws_db_event_subscription |  x  |  |
| aws_db_instance |  x  |  x  |
| aws_db_parameter_group |  x  |  |
| aws_db_proxy |  x  |  |
| aws_db_security_group |  x  |  |
| aws_db_snapshot |  x  |  x  |
| aws_db_subnet_group |  x  |  |
| aws_rds_cluster |  x  |  |
| aws_rds_cluster_endpoint |  x  |  |
| aws_rds_cluster_parameter_group |  x  |  |
| aws_rds_global_cluster |  |  |
| **redshift** |
| aws_redshift_cluster |  x  |  |
| aws_redshift_event_subscription |  x  |  |
| aws_redshift_parameter_group |  x  |  |
| aws_redshift_security_group |  |  |
| aws_redshift_snapshot_copy_grant |  x  |  |
| aws_redshift_snapshot_schedule |  x  |  |
| aws_redshift_subnet_group |  x  |  |
| **route53** |
| aws_route53_health_check |  x  |  |
| aws_route53_zone |  x  |  |
| **route53resolver** |
| aws_route53_resolver_endpoint |  x  |  x  |
| aws_route53_resolver_query_log_config |  x  |  x  |
| aws_route53_resolver_query_log_config_association |  |  x  |
| aws_route53_resolver_rule |  x  |  x  |
| aws_route53_resolver_rule_association |  |  |
| **s3** |
| aws_s3_bucket |  x  |  x  |
| **s3outposts** |
| aws_s3outposts_endpoint |  |  x  |
| **sagemaker** |
| aws_sagemaker_app_image_config |  |  x  |
| aws_sagemaker_code_repository |  |  x  |
| aws_sagemaker_endpoint |  x  |  x  |
| aws_sagemaker_feature_group |  x  |  x  |
| aws_sagemaker_model |  x  |  x  |
| aws_sagemaker_model_package_group |  x  |  x  |
| **secretsmanager** |
| aws_secretsmanager_secret |  x  |  |
| **securityhub** |
| aws_securityhub_action_target |  |  |
| aws_securityhub_insight |  |  |
| **servicecatalog** |
| aws_servicecatalog_portfolio |  x  |  x  |
| aws_servicecatalog_service_action |  |  |
| aws_servicecatalog_tag_option |  |  |
| **servicediscovery** |
| aws_service_discovery_service |  x  |  x  |
| **ses** |
| aws_ses_active_receipt_rule_set |  |  |
| aws_ses_configuration_set |  |  |
| aws_ses_domain_identity |  |  |
| aws_ses_email_identity |  |  |
| aws_ses_receipt_filter |  |  |
| aws_ses_receipt_rule_set |  |  |
| aws_ses_template |  |  |
| **sfn** |
| aws_sfn_activity |  x  |  x  |
| aws_sfn_state_machine |  x  |  x  |
| **signer** |
| aws_signer_signing_job |  |  |
| aws_signer_signing_profile |  x  |  |
| **sns** |
| aws_sns_platform_application |  |  |
| aws_sns_topic |  x  |  |
| aws_sns_topic_subscription |  |  |
| **sqs** |
| aws_sqs_queue |  x  |  |
| **ssm** |
| aws_ssm_activation |  x  |  |
| aws_ssm_association |  |  |
| aws_ssm_document |  x  |  |
| aws_ssm_maintenance_window |  x  |  |
| aws_ssm_parameter |  x  |  |
| aws_ssm_patch_baseline |  x  |  |
| aws_ssm_resource_data_sync |  |  |
| **storagegateway** |
| aws_storagegateway_gateway |  x  |  |
| aws_storagegateway_tape_pool |  x  |  |
| **synthetics** |
| aws_synthetics_canary |  x  |  |
| **timestreamwrite** |
| aws_timestreamwrite_database |  x  |  x  |
| **transfer** |
| aws_transfer_server |  x  |  |
| **waf** |
| aws_waf_byte_match_set |  |  |
| aws_waf_geo_match_set |  |  |
| aws_waf_ipset |  |  |
| aws_waf_rate_based_rule |  x  |  |
| aws_waf_regex_match_set |  |  |
| aws_waf_regex_pattern_set |  |  |
| aws_waf_rule |  x  |  |
| aws_waf_rule_group |  x  |  |
| aws_waf_size_constraint_set |  |  |
| aws_waf_sql_injection_match_set |  |  |
| aws_waf_web_acl |  x  |  |
| aws_waf_xss_match_set |  |  |
| **wafregional** |
| aws_wafregional_byte_match_set |  |  |
| aws_wafregional_geo_match_set |  |  |
| aws_wafregional_ipset |  |  |
| aws_wafregional_rate_based_rule |  x  |  |
| aws_wafregional_regex_match_set |  |  |
| aws_wafregional_regex_pattern_set |  |  |
| aws_wafregional_rule |  x  |  |
| aws_wafregional_rule_group |  x  |  |
| aws_wafregional_size_constraint_set |  |  |
| aws_wafregional_sql_injection_match_set |  |  |
| aws_wafregional_web_acl |  x  |  |
| aws_wafregional_xss_match_set |  |  |
| **wafv2** |
| aws_wafv2_web_acl_logging_configuration |  |  |
| **worklink** |
| aws_worklink_fleet |  |  x  |
| **workspaces** |
| aws_workspaces_directory |  x  |  |
| aws_workspaces_ip_group |  x  |  |
| aws_workspaces_workspace |  x  |  |
| **xray** |
| aws_xray_group |  x  |  |

## Acceptance tests

***IMPORTANT:*** Acceptance tests create real resources that might cost you money. Also, note that if you contribute a
PR, the [Travis build](https://travis-ci.org/github/jckuester/awsweeper) will always fail since AWS credentials are not
injected into the PR build coming from forks for security reasons. You can either run tests locally against your
personal AWS account or ask me to run them for you instead.

Run all acceptance tests with

    AWS_PROFILE=<myaccount> AWS_DEFAULT_REGION=us-west-2 make test-all

or to test the working of AWSweeper for a just single resource, such as `aws_vpc`, use

    AWS_PROFILE=<myaccount> AWS_DEFAULT_REGION=us-west-2 make test-all TESTARGS='-run=TestAcc_Vpc*'

## Disclaimer

You are using this tool at your own risk! I will not take responsibility if you delete any critical resources in your
production environments.

+++
title = "Alerts"
aliases = ["/docs/grafana/latest/alerting/rules/", "/docs/grafana/latest/alerting/metrics/"]
weight = 110
+++

# Grafana alerts

Alerts allow you to know about problems in your systems moments after they occur. Robust and actionable alerts help you identify and resolve issues quickly, minimizing disruption to your services. 

Grafana 8.0 has new and improved alerts. The new alerting system are an [opt-in]({{< relref "./unified-alerting/opt-in.md" >}}) feature that centralizes alerting information for Grafana managed alerts and alerts from Prometheus-compatible data sources in one UI and API.

Alerts have four main components:

- Alerting rule - One or more query and/or expression, a condition, the frequency of evaluation, and the (optional) duration that a condition must be met before creating an alert.
- Contact point - A channel for sending notifications when the conditions of an alerting rule are met.
- Notification policy - A set of matching and grouping criteria used to determine where, and how frequently, to send notifications.
- Silences - Date and matching criteria used to silence notifications.

You can create and edit alerting rules for Grafana managed alerts, Cortex alerts, and Loki alerts as well as see alerting information from prometheus-compatible data sources in a single, searchable view. For more information, on how to create and edit alerts and notifications, refer to [Overview of Grafana 8.0 alerts]({{< relref "../alerting/unified-alerting/_index.md" >}}).

As part of the new alert changes, we have introduced a new data source, Alertmanager, which includes built-in support for Prometheus Alertmanager. It is presently in alpha and it not accessible unless alpha plugins are enabled in Grafana settings. For more information, refer to [Alertmanager data source]({{< relref "../datasources/alertmanager.md" >}}).

> **Note:** Out of the box, Grafana still supports old Grafana alerts. They are legacy alerts at this time, and will be deprecated in a future release. For more information, refer to [Legacy Grafana alerts]({{< relref "./old-alerting/_index.md" >}}).

To learn more about the differences between new alerts and the legacy alerts, refer to [What's New with Grafana 8 Alerts]({{< relref "../alerting/difference-old-new.md" >}}).
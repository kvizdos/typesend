# TypeSend

TypeSend is a centralized email template management and sending service for developers. It handles email templates, offers a pre-built UI for modifying emails (and unsubscribing with one click), and tracks send history, open rates, click rates, and more. Rather than rebuilding email handling for every app, you simply drop in TypeSend and start sending emails via SendGrid (or later, SES, and other providers).

[![Go Test](https://github.com/kvizdos/typesend/actions/workflows/test.yaml/badge.svg)](https://github.com/kvizdos/typesend/actions/workflows/test.yaml)

## Features
### Centralized Template Management:
Store and manage email templates in MongoDB with a pre-built, embeddable Template Editor UI that displays all possible variables and provides live previews.

### Asynchronous Email Dispatch:
The Send() function enqueues messages in AWS SQS. A Lambda function processes the queue—filling templates, logging metadata (e.g., app ID, custom metadata like "to-user-id"), sending emails, and tracking analytics.

### Type-Safe Template Variables:
Define email templates’ variables strictly in code (with plans to leverage Go generics for compile-time safety) so that every template gets the exact data it needs.

### Provider Integration & Extensibility:
Out-of-the-box integration with SendGrid with an easy pathway for developers to extend support to other providers.

### Unsubscribe & Spam Safeguards:
Includes a pre-built UI for one-click unsubscribing. Automatically marks users as unsubscribed (or "never send" status) if they are identified as spam targets, ensuring compliance and a good sender reputation.

### Analytics & Tracking:
Query full message history (e.g., “show me everything this user received in the last X days”) via direct read access to the tracking data (wrapper functions included). Future plans include a full analytics UI.

### Terraform-Managed Deployment:
Fully customizable deployments with provided Terraform code for configuring templates, provider credentials, SQS/Lambda settings, and more.

## Architecture & Workflow
### Send() Function:

Your application calls Send() with the template ID, app ID, email headers (subject, recipients, etc.), and strictly defined variables.
The function packages this data into a JSON message and enqueues it into an AWS SQS queue.

### SQS Message Processing:

A Lambda function, triggered by new messages on SQS, retrieves the message.
It loads the correct template from MongoDB, renders the email by inserting the provided variables, and logs the message details (including custom metadata).
The email is then sent via the configured provider (initially SendGrid).
Open rates, click rates, and spam reports are tracked. If a recipient is identified as spammed, they are automatically marked to never receive future emails.

## Template & Unsubscribe UI:

Template Editor: A shared, embeddable UI endpoint lets developers edit and preview email templates, showing all possible variables.
Unsubscribe UI: A pre-built one-click unsubscribe page is provided, so recipients can easily opt-out. This page ensures that once a user unsubscribes or is flagged as spam, emails will no longer be sent to that address.
Analytics Querying:

For use cases like “show me everything this user received in the last X days,” you can directly query a read-only replica of your MongoDB (or a reporting database) to retrieve historical email data.

---
# Page settings
layout: default
keywords: kat,postgres,database,cli,migrations,sql
title: Configuration
description: |
    Kat is a PostgreSQL database migration tool. It allows you run your migrations with raw SQL files.
comments: false
permalink: /config/
page_nav:
    prev:
        content: Installation
        url: '/install'
    next: 
        content: Migration
        url: '/migration'
---

Kat requires a configuration file to know where to look for migrations and to know which database to perform migrations on.

```yaml
migration:
    tablename: migration_logs
    directory: /Users/bolaji/Desktop/kat-test/migrations
database:
    url: postgres://bolaji:andela@localhost:5432/katest
```

# Configuration Guide for Kat

After installing Kat, the next step is to configure it for your specific environment and requirements. This guide provides detailed instructions on how to configure the tool effectively.

## Overview

Before configuring Kat, it's important to understand the key components that require configuration:

- Database Connections: Setting up source and target database connections.
- Migration Settings: Configuring migration options, such as data mapping, filtering, and scheduling.
- Security Settings: Ensuring secure data transfer and access control.

## Database Connections

### Setting Up Source Database

1. Open [Your Tool Name] and navigate to the 'Database Connections' section.
2. Click on 'Add New Source Database'.
3. Provide the following details:
   - Database Type: [e.g., MySQL, PostgreSQL]
   - Hostname: [Source database hostname]
   - Port: [Source database port]
   - Username: [Database username]
   - Password: [Database password]
4. Click 'Test Connection' to verify the details.
5. Save the connection.

### Setting Up Target Database

1. In the 'Database Connections' section, click on 'Add New Target Database'.
2. Repeat steps 3-5 as above, providing details for your target database.

## Migration Settings

### Data Mapping

1. Navigate to 'Data Mapping' under the 'Migration Settings'.
2. Select source and target databases.
3. Map source tables to target tables.
4. [Any additional mapping options, e.g., column mapping, data type conversion]

### Filtering Data

1. In 'Data Filtering', specify criteria to include or exclude specific data.
2. [Details on how to add filters, e.g., by date, by row count]

### Scheduling Migrations

1. Go to 'Migration Scheduling'.
2. Set up a migration schedule [Details on how to schedule, e.g., one-time, recurring].

## Security Settings

### Setting Up Encryption

1. Navigate to 'Security Settings'.
2. Enable data encryption during transfer [Specific instructions].

### Access Control

1. Set up user roles and permissions [Details on how to configure user access].

## Saving and Applying Configuration

1. After configuring the settings, click 'Save Configuration'.
2. Apply the configuration by clicking 'Start Migration' or 'Apply'.

## Troubleshooting

For any issues during the configuration process, refer to our [troubleshooting guide](Insert link to troubleshooting guide) or contact [support contact details].

## Next Steps

With the configuration complete, you are now ready to start your database migration. Proceed to the [Migration section](Insert link to migration section) for detailed steps on executing your database migration.

---

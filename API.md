# Uptime Kuma API Documentation

## Overview
The Uptime Kuma API provides a comprehensive set of endpoints for managing monitoring services, status pages, maintenance windows, and user authentication. Base URL: `/api/v1`

## Authentication
- **OAuth2 Password Grant Flow**: `/login/access-token` (POST)
- All authenticated endpoints require Bearer token

## Core Resources

### Monitors
- **GET `/monitors`** - List all monitors
- **POST `/monitors`** - Create new monitor
- **GET `/monitors/{monitor_id}`** - Get monitor details
- **PATCH `/monitors/{monitor_id}`** - Update monitor
- **DELETE `/monitors/{monitor_id}`** - Delete monitor
- **POST `/monitors/{monitor_id}/pause`** - Pause monitoring
- **POST `/monitors/{monitor_id}/resume`** - Resume monitoring
- **GET `/monitors/{monitor_id}/beats`** - Get monitor heartbeats
- **POST/DELETE `/monitors/{monitor_id}/tag`** - Manage monitor tags

### Status Pages
- **GET/POST `/statuspages`** - List/create status pages
- **GET/POST/DELETE `/statuspages/{slug}`** - Manage specific status page
- **POST `/statuspages/{slug}/incident`** - Post incident to status page
- **DELETE `/statuspages/{slug}/incident/unpin`** - Unpin incident

### Maintenance
- **GET/POST `/maintenance`** - List/create maintenance windows
- **GET/PATCH/DELETE `/maintenance/{maintenance_id}`** - Manage maintenance
- **POST `/maintenance/{maintenance_id}/pause`** - Pause maintenance
- **POST `/maintenance/{maintenance_id}/resume`** - Resume maintenance
- **GET/POST `/maintenance/{maintenance_id}/monitors`** - Manage monitors in maintenance

### Tags
- **GET/POST `/tags`** - List/create tags
- **GET/DELETE `/tags/{tag_id}`** - Manage specific tag

### Users
- **GET/POST `/users`** - List/create users
- **GET/DELETE `/users/{username}`** - Manage specific user

### System Information
- **GET `/info`** - General system information
- **GET `/cert_info`** - Certificate information
- **GET `/ping`** - Average ping statistics
- **GET `/uptime`** - System uptime
- **GET `/database/size`** - Database size
- **POST `/database/shrink`** - Shrink database
- **POST `/settings/upload_backup`** - Upload backup

## Monitor Types
HTTP, port, ping, keyword, GRPC, DNS, Docker, push, Steam, GameDig, MQTT, SQL Server, PostgreSQL, MySQL, MongoDB, RADIUS, Redis

## Authentication Methods
Basic, NTLM, mTLS
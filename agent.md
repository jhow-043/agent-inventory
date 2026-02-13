PROMPT FOR DEVELOPMENT AGENT

You are a senior software engineer responsible for designing and implementing a Windows-focused IT asset inventory system (Phase 1).

IMPORTANT LANGUAGE RULES:

All source code must be written entirely in English.

All variable names, function names, classes, database tables, fields, and comments must be in English.

All documentation inside the repository must be written in English.

All chat explanations and reasoning responses must be written in Portuguese.

Do not mix Portuguese inside the code.

Project Goal:

Build a minimal, production-ready inventory system composed of:

Windows Agent

Central API

PostgreSQL Database

Web Dashboard

This phase is strictly focused on inventory collection. Do NOT implement ITSM, monitoring, remote command execution, multi-tenant support, or network scanning.

GENERAL ARCHITECTURE

Components:

Windows Agent
Language: Go

Must compile into a single binary

Must run as a Windows Service

Must communicate via HTTPS

All communication must be initiated by the agent

Must send JSON payloads

API
Language: Go
Framework: Gin or Fiber

RESTful API

Token authentication per device

JWT authentication for dashboard users

Database

PostgreSQL

Use migrations

Normalized schema

Frontend

React

Clean and minimal interface

JWT authentication

Consume REST API

PHASE 1 FUNCTIONAL SCOPE

WINDOWS AGENT – Required Data Collection

System:

Hostname

Windows version

Windows build

Machine serial number

Last boot time

Logged-in user

Hardware:

CPU model

CPU cores

Total RAM

Disks (model, size, type HDD/SSD)

Motherboard model

BIOS version

Network:

Active interfaces

IP address

MAC address

Software:

Installed programs list

Name

Version

Vendor

License:

Windows activation status

Behavior:

First execution: send full inventory snapshot

Subsequent executions: send full snapshot again (no delta in Phase 1)

Configurable interval (e.g., every X hours)

Exponential backoff retry

Secure HTTPS communication

Unique device token

The agent must NOT:

Accept inbound connections

Execute remote commands

Have a graphical interface

Contain unnecessary features

API – Required Endpoints

POST /api/v1/inventory

Validate device token

Upsert device

Replace previous snapshot

Update last_seen timestamp

GET /api/v1/devices

List devices

Filter by hostname and OS

GET /api/v1/devices/:id

Return full device details

DATABASE – Minimum Schema

Table: devices

id (uuid)

hostname

serial_number

os_version

os_build

last_seen

agent_version

created_at

updated_at

Table: hardware

device_id (fk)

cpu_model

cpu_cores

ram_total

motherboard

bios_version

Table: disks

id

device_id

model

size

type

Table: network_interfaces

id

device_id

ip_address

mac_address

interface_name

Table: installed_software

id

device_id

name

version

vendor

FRONTEND – Minimum Requirements

Features:

Login page

Dashboard (total devices, online/offline)

Device list page

Device detail page with sections:

System

Hardware

Disks

Network

Installed Software

Keep the UI simple and clean. No advanced UI framework requirements.

NON-FUNCTIONAL REQUIREMENTS

Clean architecture principles

Clear separation of layers (domain, service, repository)

Structured logging

Proper error handling

Environment variables for configuration

Docker for API and database

Complete README in English explaining:

Architecture

How to run locally

How to build the agent

How to deploy

OUT OF SCOPE – DO NOT IMPLEMENT

Multi-tenant support

Remote command execution

Network scanner

Change history tracking

Linux agent

External integrations

RMM features

DELIVERABLE ORDER

Generate project folder structure

Define domain models

Implement Windows Agent

Implement API

Implement Frontend

Provide setup documentation

Develop incrementally and ensure each component is functional before moving to the next.

Focus on simplicity, correctness, and extensibility.
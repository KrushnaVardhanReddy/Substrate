# P9-T06: Configuration Management (Viper)

## Overview
Migrate all `os.Getenv` calls to Viper for robust `.env`, CLI flag, and YAML configuration loading.

## Requirements
1. **Viper Setup**: Initialize Viper in the CLI root command to load from `.env`, environment variables, and `substrate.yaml`.
2. **Migration**: Replace every `os.Getenv("...")` call in the codebase with `viper.GetString("...")`.
3. **Flag Binding**: Bind all Cobra CLI flags to Viper keys.
4. **Precedence**: Environment variables must override YAML config; CLI flags must override environment variables.

# Grub Configurator

A simple GUI application for configuring the GRUB bootloader, aimed for Arch Linux systems.

## Features

- **Grub Config Editor:** A simple text editor to modify the `/etc/default/grub` file.
- **Boot Order Manager:** A simple list to reorder the boot entries.
- **Theme Manager:** A simple GUI to set the GRUB theme.
  - Drag and drop a theme folder to set it as the GRUB theme.
  - Select a theme folder using a file dialog.
- **Safe and Reliable:** Uses the recommended way of modifying the GRUB bootloader by creating a custom script in `/etc/grub.d/`.

## Screenshots

*A screenshot of the application will be added here soon.*

## Building and Running

To build and run the application, you need to have Go and Fyne installed.

```bash
# Tidy the dependencies
go mod tidy

# Run the application
go run .
```

## Disclaimer

This application modifies the GRUB bootloader configuration. A mistake can render your system unbootable. Please use it with caution and at your own risk. It is highly recommended to back up your GRUB configuration before making any changes.
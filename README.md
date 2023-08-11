
# pkgparser Go Package

## Overview

custom parser for `go-dpkg` , `go-rpm`, `go-packman` packages

## Data Structures

### Package

Contains detailed information about a Debian package. Key fields include:

- `PackageName`: Name of the package.
- `Status`, `Priority`, `Section`, and more: Various details about the package.
- `Maintainer`, `OriginalMaintainer`: Contact information of the package maintainers.
- `Extra`: A map that can hold any additional information about the package not covered by the other fields.

### PackageContact

Represents the contact information of a package. It can be either an email address or a website.

- `Name`: The name of the contact.
- `Contact`: The actual contact information - email or website.
- `Type`: Specifies if the contact is an "email" or "website".

## License

Refer to the `LICENSE` file in the repository.

## Contributing

Contributions are welcome! Please submit a pull request or open an issue on the project's repository.
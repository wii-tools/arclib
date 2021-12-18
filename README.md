# arclib
> **A library to read and write U8 archives.**
>
> [Documentation](https://pkg.go.dev/github.com/wii-tools/arclib)

## License
**arclib** is released under the MIT License, read [here](/LICENSE) for more information.

## Q&A
### This library is an absolute mess.
Thanks!

### Why does this library not utilize [fs.FS](https://pkg.go.dev/io/fs#FS)?
Unfortunately, the `fs.FS` package only provides read-only functions.
There are no write capabilities, making it unfortunately unsuitable. Implementing read-only and custom write feels hacky.
In the future, it would be highly preferable to migrate this library to utilize its interface.
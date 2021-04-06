# 🦉 Docuowl

Docuowl generates a static single-page documentation from Markdown files

## Rationale
As a long-time fan of documentation style made by [Stripe](https://stripe.com/docs/api),
and Markdown, I decided to use the former as a base to create a pretty documentation
generator that outputs something like Stripe's. Stripe also generously allowed me
to use their layout, so here's a big thank you to them! ♥️

## Demo
Looking for a demo? A simple demo is available at [https://docuowl.github.io/demo/](https://docuowl.github.io/demo/)!

<center><img width="1676" alt="image" src="https://user-images.githubusercontent.com/77198/113536836-69b0b000-95ad-11eb-9003-f3bf6a2e3f9f.png">
</center>

## Documentation Organization
Docuowl takes a directory as input. The directory is expected to have one
directory for each section or group. Each group may have subsections, which by
their turn must also be placed into directories.
Each **Section** is required to have an `content.md` file, containing the
Frontmatter for that section, and an optional `sidenotes.md` file, that will be
rendered to the right of the section. The Frontmatter must contain at least a
`Title` property, and an optional `ID` property containing a unique slug for that
section.
Each **Group** must contain a single `meta.md` file, containing a Frontmatter like
a Section, and an optional content following the frontmatter.

For instance, take the following directory tree as example:

```
.
├── 1-introduction
│   └── content.md
├── 2-errors
│   ├── content.md
│   └── sidenotes.md
├── 3-authentication
│   ├── content.md
│   └── sidenotes.md
├── 4-authorization
│   ├── 1-login
│   │   ├── content.md
│   │   └── sidenotes.md
│   ├── 2-logout
│   │   ├── content.md
│   │   └── sidenotes.md
│   ├── 4-me
│   │   ├── content.md
│   │   └── sidenotes.md
│   └── meta.md
├── 5-foo
│   ├── 1-listing-foos
│   │   ├── content.md
│   │   └── sidenotes.md
│   ├── 2-merged-foos
│   │   ├── content.md
│   │   └── sidenotes.md
│   └── meta.md
├── 6-bars
│   ├── content.md
│   └── sidenotes.md
├── 7-list-foobars
│   ├── content.md
│   └── sidenotes.md
├── 8-get-foobar
│   ├── content.md
│   └── sidenotes.md
└── 9-foobar-data
    ├── content.md
    └── sidenotes.md
```

### Example of `meta.md`:

```markdown
---
Title: Authorization
---

> :warning: **Warning**: All authorization endpoints are currently in maintenance
```

### Markdown Extensions

Docuowl introduces two new blocks to Markdown: Boxes and Attributes List.

#### Boxes
Boxes can only be used in sidenotes. To create a new box, use the following
format:

```
#! This is a box
And this is the box's content
```

After one `#!`, the box will take any content that follows until one of the
following conditions are met:

1. A horizontal ruler is found (`----`)
2. Another Box begins.

#### Attributes List
Attributes Lists can only be used in contents. To create a new Attribute List,
use the following format:

```
#- Attribute List
- Key1 `type`
- Key1 Description
```

## Usage
Docuowl can be invoked in two modes: Compile, and Watch.

### Compile
Compilation will output a single `index.html` file to an specified directory,
taking another directory as input. For instance:

```bash
$ docuowl --input docs --output docs-html
```

### Watch
Watch allows one to continuously write documentation and see the preview with
auto-reload. For that, use:

```bash
$ docuowl --input docs --output docs-html --watch

Docuowl v0.1
Listening on 127.0.0.1:8000
```

Then open your browser and point to 127.0.0.1:8000. The page will be reloaded
each time a file changes in the `input` directory.

## TODO
- [ ] Full-text Search
- [ ] Improve CSS

## License

This software uses other open-source components. For a full list, see the `LICENSE` file.

```
MIT License

Copyright © 2021 Victor Gama
Copyright © 2021 Real Artists

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

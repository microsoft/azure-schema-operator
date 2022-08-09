# Contributing to the Documentation

The documentation site is hosted on [GitHub Pages](https://pages.github.com)  
The site is generated using [Hugo](https://gohugo.io)  

## Pre-requisits

To work on the decumentation you will need:

1. a text file editor, perferably with Markdown support.
1. Hugo installed locally to test the changes.
1. Task - to run the taskfile tasks.

## How-To

The documentation sits in the `docs/site` folder.  
The content sits in the `contenet` folder.
and static files such as images and samples reside in the `static` folder.  

To view local changes before pushing, run:

```bash
task run-docs-site
```

This will run Hugo locally and open them in a new Tab.

## Helm Chart Documentation

If changes where made to the Helm chart parameters, we also need to update the helm chart documentation.

To update them, run:

```bash
task helm-docs
```

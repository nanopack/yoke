# Yoke Site HTML

This directory contains a [Middleman](https://middlemanapp.com/) app used to generate all the HTML for [nanopack.io/yoke](http://nanopack.io/yoke). This includes the Yoke landing page and documentation.

## Contributing

Yoke is a [nanopack project](http://nanopack.io) and all contributions should follow the [Nanopack Contribution Guidelines](http://nanopack.io/contributing/).


### Run Nanobox Dev
[Nanobox](https://desktop.nanobox.io) is a local development tool that sets up an isolated virtualized environment in which you can modify the site code. If you have not already, [download & install nanobox](https://desktop.nanobox.io/downloads/), then `cd` into the `site` directory and run:

```bash
nanobox dev
```

This will provision a VM, install all the necessary Middleman dependencies, run and modify your local hosts file to make the dev site available at a unique URL, then drop you into an interactive console. Once in the console, run:

```bash
bundle exec middleman server
```

The dev site will then be available at the URL provided by nanobox and the code in your local filesystem will be watched and live reload the dev site when modified.

### Updating Content
Any updates to content should be made in the `source` directory unless modifying content on the main landing page or the documentation index. This are generated from yaml files inside of the `data` directory.

#### Landing Page Content
The Yoke landing page content is generated from the `project_info.yml` file inside of the `data` directory.

#### Updating Documentations
Yoke documentation is housed inside of the the `source/docs` directory. Docs should be written in Markdown ([GFM](https://help.github.com/articles/github-flavored-markdown/) specifically). Middleman parses files based on the file extension(s) and extension order. All docs should be saved as `doc-name.html.md`. This will run it through the Markdown parser first, then generate the HTML for the apge.

Docs can be nested inside of subdirectories, but should not be nested more than 3 layers deep. Every subdirectory needs to include an `index.html.md` which contains the content of the the landing page for that directory.

#### Documentation Index
The docs index is generated from the `data/docs_index.yml` file. Again, docs should not be nested more than 3 layers deep. In order for new docs to show up in the navigation, the must be added to the `doc_index.yml` data file.

## Commit Your Code, Push, & Submit a Pull Request
When you're ready to push your changes, simply commit your code, push to your forked repo, then create a pull request on the Yoke project.



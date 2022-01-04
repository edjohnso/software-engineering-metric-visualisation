# Torvalds Number

![GitHub CI Workflow](https://github.com/edjohnso/software-engineering-metric-visualisation/actions/workflows/ci.yaml/badge.svg)

A simple project to demonstrate accessing and visualising data from the GitHub API.

I chose to invent my own software engineering metric to measure and visualise: The number of repositories between open-source developers.
If someone else owns a repository that you have contributed to, then I deem you a collaborator with the other individual. If the other individual has
then contributed to the repository of a third person, then I deem them to be a collaborator with the third person and you to have a Torvalds Number of 2 with the
third person. As such, this metric counts the distance in collaboration between individuals. Additionally, trivial metrics can be compared between collaborators.
I was hoping to spend more time on these traditional metrics so I could try look for some correlation.
Maybe there's a relationship between proximity in terms of collaboration with Linus Torvalds and software engineering productivity? Who knows? Well, this tool can at least help with that question.

I implemented this project with a custom Go webserver and pure JavaScript with the HTML Canvas element. I mainly focused on perfecting my webserver with
API request caching and collaborators graph generation. I have learned a lot about different areas of web development for this full-stack project. I also
spent a decent amount of time experimenting with development tools such as Docker and the suite of Go testing tools. I spent a while learning about WebSockets
so that the Go backend webserver could send updates to the JavaScript frontend while the frontend could send commands back to the server.

![Screenshot](/.github/screenshot.png)

## Getting Started

Below is a rundown on how to fully get setup and running. The TL;DR is:
 - Register a GitHub OAuth App with a callback address set to `http://localhost`.
 - Create a GitHub Personal Access Token (only used for unit tests).
 - Export the GitHub OAuth App Client ID and Secret and the Personal Access Token as the environment variables GHO_CLIENT_ID, GHO_CLIENT_SECRET and GHO_PAT.
 - Execute `make run` to start the webserver and then visit `http://localhost` in your web browser.

### Dependencies

If you want to host this on your own physical machine, you will need a Go compiler. Optionally, you may also need GNU Make.
Alternatively, if you have a Docker daemon running, I have provided a Dockerfile which can build and run the webserver containerized.

You will also need an internet connection to access the GitHub API.

### API Keys

Before you can even build the project, you are going to need to get a few secrets from GitHub. First, you must register a new GitHub OAuth App.
This is to allow the webserver to access the GitHub API under your name. Most of the options don't matter, but the callback address should be set
to `http://localhost` if you want to test the server out on your machine.
[You can find more information here](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app).

Additionally, if you are planning to run the unit tests, you will need to pick up a GitHub Personal Access Token with an empty scope.

For the webserver to use these secret keys, you must store them in three special environment variables when start it.
The GitHub OAuth Client ID should be stored in the GHO_CLIENT_ID environment variable, the Client Secret should be stored in the GHO_CLIENT_SECRET
environment variable and the Personal Access Token should be stored in the GHO_PAT environment variable. It is up to you how you accomplish this
but personally I put all my secrets in a `secrets.env` file and use that with the `env` command to populate the environment variables just when
I'm building or running the server. The might look something like `env $(cat secrets.env) make run`.

### Building & Running

![GIF animation of the Docker image being build and run](/.github/docker.gif)

You can either start the webserver as a Docker container or run it straight on your machine.
Both use-cases are covered by the provided Makefile. Assuming the API keys required are
stored correctly as environment variables, you can just run `make docker` or `make run`.
Both of these automatically run the unit tests and the Go Vet tool.

You could also just use the Go compiler with `go build ./cmd/webserver` and then execute
the compiled binary with `./webserver 8080 ./web/public ./web/templates/* cache.gz`.

## Usage

### Accessing the webpage

![GIF animation of the login process](/.github/login.gif)

Once you've launched the server, you can visit `http://localhost` to get started.
There, you should see the login page. If you've correctly setup GitHub OAuth for this
application, then clicking on the *Sign In With GitHub* button should bring you to
an official GitHub authorization page. Again assuming you correctly specified `localhost`
as the callback address, upon authorizing the app you will be returned to web page.
The server will have been provided your authentication token to use the API under
your name. As such, you should arrive at the main webpage.

### Using the webpage

![GIF animation of using the webpage](/.github/graph.gif)

To begin with, it will be pretty bare. You can click on the buttons in the bottom-left corner
to control searching for collaborators and drawing deeper nodes on the graph once you've gone
a few layers down.
The arrow keys can be used to pan the viewpoint around.

Licensed under GPLv3\
Ted Johnson 2021

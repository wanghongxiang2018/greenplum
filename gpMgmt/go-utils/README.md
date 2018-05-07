# Utilities written in golang

## Fetching dependencies

We are vendoring dependencies with go dep. You can install dependencies for a
given utility by running this command in the utility directory:

	$ dep ensure

Please see the README for a particular utility for more information, at
src/<utility_name>/README.md

## Development

Ensure you have Golang installed. You can do this easily on macOS with brew:

	$ brew install go

For our Go utilities, we are using a non-standard $GOPATH: the gpMgmt/go-utils
sub-directory of gpdb source. When developing with these utilities, set your
GOPATH and PATH variables as follows:

	export GOPATH=$HOME/workspace/gpdb/gpMgmt/go-utils
	export PATH=$HOME/workspace/gpdb/gpMgmt/go-utils/bin:$PATH

### Support for IDE development

If you are using an IDE, depending on the IDE, you may need to provide it with
the custom GOPATH by adding the above lines to your .bashrc or .bash_profile
file.

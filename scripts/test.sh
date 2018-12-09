
cd tests/go

go test -v

cd ../..

source ./scripts/build.sh
source ./tests/sh/helpers.sh

# source ./tests/sh/test_simple.sh --package
# source ./tests/sh/test_rumors.sh --package
source ./tests/sh/test_files.sh --package
source ./tests/sh/test_private.sh --package
source ./tests/sh/test_routing.sh --package
source ./tests/sh/test_search.sh --package
source ./tests/sh/test_download.sh --package
source ./tests/sh/test_chain.sh --package

print_test_results

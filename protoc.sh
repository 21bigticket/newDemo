protoc --go_out=. --go_opt=paths=source_relative --go-triple_out=. --go-triple_opt=paths=source_relative ./api/samples_api.proto && \
if [ -f samples_api.triple.go ]; then
  mv samples_api.triple.go api/
fi
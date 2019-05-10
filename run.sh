rm compile.app
cd go
go build -o compile.app
./compile.app $@
cd ..

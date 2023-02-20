go build miniServer.go
echo "build success"
echo ""

echo "try move to : /usr/bin/miniServer"
sudo mv miniServer /usr/bin/
echo ""

echo "install success"
echo ""
echo "version message: "
/usr/bin/miniServer -v
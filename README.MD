# River SDK
This is the SDK which could be used by GUI/CLI clients to connect to River (https://river.im) server.
There is also an CLI for debugging and testing purpose in cmd/cli-river.

We have two versions of the SDK, one is the complete solution named 'prime' and the second one 'mini'
is useful when we don't need the full functionality.

# How TO Build for Mobiles
for iOS:
run `sh build-ios-prime.sh`
run `sh build-ios-mini.sh`

for Android:
run `sh build-android-prime.sh`
run `sh build-android-mini.sh`


# Main Components
We have four important sub package, we call them controllers.
1. File Controller
2. Network Controller
3. Queue Controller
4. Sync Controller


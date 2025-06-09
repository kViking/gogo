# GoGoGadget

GoGoGadget lets you run your favorite PowerShell commands with easy shortcuts. You don't need to know anything about programming or scripts—just use simple commands in your terminal!

---

## What is GoGoGadget?

GoGoGadget is like a command line Swiss Army knife that you build for yourself! It's is a tool that helps you:

- Save your favorite PowerShell commands as shortcuts
- Run those shortcuts quickly, filling in any details you need
- List all your saved shortcuts
- Remove shortcuts you don't need anymore

You don't need to install anything except GoGoGadget. The installer will set everything up for you.

---

## How Do I Use GoGoGadget?

After installing, open a new terminal or command prompt and use these commands:

>### You Can Always Ask for Help
>
>Every command and subcommand in GoGoGadget has accessible, easy-to-read help. Just add `--help` (or `-h`) to any command to see what it does and how to use it.

### 1. Add a New Shortcut

Type:

```powershell
GoGoGadget add
```

GoGoGadget will ask you:

- What PowerShell command do you want to save? (You can use `{{VARNAME}}` for things you want to fill in each time)
- What name do you want to give this shortcut?
- What does this shortcut do? (A short description)
- What does each variable mean? (If you used any)

**Example:**

```powershell
PS C:\Users\You> GoGoGadget add
📝 Enter PowerShell command (use {{VARNAME}} for variables): (Get-ChildItem {{folder}} -Recurse -File | Measure-Object).Count
🔖 Enter script name: filecount
💡 Enter script description: Counts all files in a directory recursively (all the files in all the folders)
✏️  Describe variable 'folder': Folder to count

PS C:\Users\You> GoGoGadget filecount .\Documents
1239
```

### 2. See All Your Shortcuts

Type:

```powershell
GoGoGadget list
```

This will show you all the shortcuts you have saved (and only the ones you saved).

### 3. Run a Shortcut

Type:

```powershell
GoGoGadget greet -Name "Alice"
```

Or, if your shortcut only has one variable, you can just type:

```powershell
GoGoGadget greet "Alice"
```

This will run your shortcut and fill in the variable with what you typed.

### 4. Delete a Shortcut

Type:

```powershell
GoGoGadget delete greet
```

This will remove the shortcut called `greet`.

---

## Analyze Your PowerShell Commands

GoGoGadget 0.1.1 introduces the `analyze` command! If you’re not sure which parts of your PowerShell command should be variables, just run:

```powershell
gogogadget analyze --command "Get-Content 'C:\\Users\\me\\file.txt' -Encoding UTF8"
```

Or simply:

```powershell
gogogadget analyze
```

and paste your command when prompted. GoGoGadget will highlight likely user input sections (like file paths, numbers, or strings) and suggest how to turn them into variables. **The analyzer is experimental and only provides suggestions—please review and adjust as needed for your use case.** This makes creating shortcuts even easier for everyone!

---

## Need Help?

If you type a command wrong, GoGoGadget will show you what to do. You can always see your shortcuts with:

```powershell
GoGoGadget list
```

---

## About

GoGoGadget is free and open source. You can use it for anything. If you have questions, ask the person who gave you the installer. Report any issues on our [github](https://github.com/kViking/gogo/issues) and check our [releases](https://github.com/kViking/gogo/releases) page for new versions!

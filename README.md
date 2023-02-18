# aggro

A small utility for running npm scripts.

## What it do?

This tool make it easy to run npm scripts spread over different directories in
parallel. Output of the scripts will be shown side by side in the console.

## Why do I care?

You might care, if you want to run some (eg. `watch`) scripts in parallel
without reorganizing your code or writing a custom script.

## Usage

    aggro <pattern-1> ... <pattern-n> <basedir>
    
    Where
        <pattern> - a name for a scripts block in package.json (eg. 'watch:local')
        <basedir> - directory inside which to look (recursively) for package.json files

## Example

`aggro watch:sass watch:js ~/src/` will recursively search all `package.json`
files in `~/src` for scripts with name `watch:sass` or `watch:js`. It will then
run all found scripts, showing the respective outputs side by side in the
console.

## License

Lincensed under European Union Public Licence. See LICENSE file.

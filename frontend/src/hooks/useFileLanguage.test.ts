import { describe, it, expect } from 'vitest';
import { getLanguageForFile } from './useFileLanguage';

describe('getLanguageForFile', () => {
  it('returns typescript for .ts files', () => {
    expect(getLanguageForFile('main.ts')).toBe('typescript');
  });

  it('returns typescript for .tsx files', () => {
    expect(getLanguageForFile('App.tsx')).toBe('typescript');
  });

  it('returns javascript for .js files', () => {
    expect(getLanguageForFile('index.js')).toBe('javascript');
  });

  it('returns javascript for .jsx files', () => {
    expect(getLanguageForFile('Component.jsx')).toBe('javascript');
  });

  it('returns python for .py files', () => {
    expect(getLanguageForFile('app.py')).toBe('python');
  });

  it('returns go for .go files', () => {
    expect(getLanguageForFile('main.go')).toBe('go');
  });

  it('returns rust for .rs files', () => {
    expect(getLanguageForFile('lib.rs')).toBe('rust');
  });

  it('returns json for .json files', () => {
    expect(getLanguageForFile('package.json')).toBe('json');
  });

  it('returns html for .html files', () => {
    expect(getLanguageForFile('index.html')).toBe('html');
  });

  it('returns css for .css files', () => {
    expect(getLanguageForFile('styles.css')).toBe('css');
  });

  it('returns scss for .scss files', () => {
    expect(getLanguageForFile('theme.scss')).toBe('scss');
  });

  it('returns markdown for .md files', () => {
    expect(getLanguageForFile('README.md')).toBe('markdown');
  });

  it('returns yaml for .yml files', () => {
    expect(getLanguageForFile('docker-compose.yml')).toBe('yaml');
  });

  it('returns yaml for .yaml files', () => {
    expect(getLanguageForFile('config.yaml')).toBe('yaml');
  });

  it('returns shell for .sh files', () => {
    expect(getLanguageForFile('build.sh')).toBe('shell');
  });

  it('returns sql for .sql files', () => {
    expect(getLanguageForFile('schema.sql')).toBe('sql');
  });

  it('returns c for .c files', () => {
    expect(getLanguageForFile('main.c')).toBe('c');
  });

  it('returns cpp for .cpp files', () => {
    expect(getLanguageForFile('main.cpp')).toBe('cpp');
  });

  it('returns c for .h header files', () => {
    expect(getLanguageForFile('util.h')).toBe('c');
  });

  it('returns dockerfile for Dockerfile', () => {
    expect(getLanguageForFile('Dockerfile')).toBe('dockerfile');
  });

  it('returns dockerfile case-insensitively', () => {
    expect(getLanguageForFile('dockerfile')).toBe('dockerfile');
  });

  it('returns makefile for Makefile', () => {
    expect(getLanguageForFile('Makefile')).toBe('makefile');
  });

  it('returns makefile case-insensitively', () => {
    expect(getLanguageForFile('makefile')).toBe('makefile');
  });

  it('returns plaintext for .txt files', () => {
    expect(getLanguageForFile('notes.txt')).toBe('plaintext');
  });

  it('returns plaintext for .env files', () => {
    expect(getLanguageForFile('.env')).toBe('plaintext');
  });

  it('returns plaintext for unknown extensions', () => {
    expect(getLanguageForFile('data.xyz')).toBe('plaintext');
  });

  it('returns plaintext for .gitignore', () => {
    expect(getLanguageForFile('.gitignore')).toBe('plaintext');
  });

  it('is case-insensitive for extensions', () => {
    expect(getLanguageForFile('App.TSX')).toBe('typescript');
  });

  it('handles files with multiple dots', () => {
    expect(getLanguageForFile('my.component.tsx')).toBe('typescript');
  });

  it('returns java for .java files', () => {
    expect(getLanguageForFile('Main.java')).toBe('java');
  });

  it('returns ruby for .rb files', () => {
    expect(getLanguageForFile('app.rb')).toBe('ruby');
  });

  it('returns php for .php files', () => {
    expect(getLanguageForFile('index.php')).toBe('php');
  });

  it('returns graphql for .graphql files', () => {
    expect(getLanguageForFile('schema.graphql')).toBe('graphql');
  });

  it('returns xml for .xml files', () => {
    expect(getLanguageForFile('pom.xml')).toBe('xml');
  });

  it('returns toml for .toml files', () => {
    expect(getLanguageForFile('Cargo.toml')).toBe('toml');
  });
});

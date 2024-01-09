{ pkgs, config, inputs, lib, ... }:

{
    languages.javascript.enable = true;
    languages.javascript.corepack.enable = true;

    scripts.install.exec = ''
        pnpm --prefix api install
        pnpm --prefix frontend install
        '';
}
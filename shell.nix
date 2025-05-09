{ pkgs ? import <nixpkgs> {} }:

(pkgs.buildFHSUserEnv {
  name = "jellyfin-tui";
  targetPkgs = pkgs: with pkgs; [
    go
    mpv
    yt-dlp
  ];
  profile = ''
    export GO111MODULE=on
  '';
}).env
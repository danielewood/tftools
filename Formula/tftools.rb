# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Tftools < Formula
  desc "Easy CLI with useful terraform/terragrunt tools"
  homepage "https://github.com/containerscrew/tftools"
  version "0.6.0"
  license "Apache 2.0 license"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/containerscrew/tftools/releases/download/v0.6.0/tftools-darwin-arm64.tar.gz"
      sha256 "50ba8ef610d03183432dbc7b6af3c12d08779db1ad13f21bb6b80f110ea1d5be"

      def install
        bin.install "tftools"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/containerscrew/tftools/releases/download/v0.6.0/tftools-darwin-amd64.tar.gz"
      sha256 "156f47166c2f2d5ed4890f5b7951e26fbb2ce8ddad6c9e02de30f24c9da8c1c6"

      def install
        bin.install "tftools"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/containerscrew/tftools/releases/download/v0.6.0/tftools-linux-arm64.tar.gz"
      sha256 "8f9e69c93db47191d07634a31c384ee55331f46d1c48056f231dafef22af4cf6"

      def install
        bin.install "tftools"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/containerscrew/tftools/releases/download/v0.6.0/tftools-linux-amd64.tar.gz"
      sha256 "ca1206562b71e8e3215a342ceecc5042c02d463924d5a61dd2c552d8403f673a"

      def install
        bin.install "tftools"
      end
    end
  end
end

#!/usr/bin/pwsh

New-Item -ItemType Directory -Force -Path @(
	"build/windows_x64",
	"build/windows_x86",
	"build/linux_x64",
	"build/linux_x86",
	"build/darwin_x64",
	"build/darwin_x86",
	"build/darwin_arm64"
) | Out-Null

$rootDir = Get-Location
Write-Host "rootDir: $rootDir"

if ($IsWindows) {
	$script:GoExe = "go.exe"
} else {
	$script:GoExe = "go"
}

class BuildTarget {
	[String]$OS
	[String]$Arch
	[String]$BuildDir

	BuildTarget(
		[String]$OS,
		[String]$Arch,
		[String]$BuildDir
	) {
		$this.OS = $OS
		$this.Arch = $Arch
		$this.BuildDir = $BuildDir
	}
}

$targets = @(
	[BuildTarget]::new("linux", "amd64", "linux_x64"),
	[BuildTarget]::new("linux", "386", "linux_x86"),
	[BuildTarget]::new("windows", "amd64", "windows_x64"),
	[BuildTarget]::new("windows", "386", "windows_x86"),
	[BuildTarget]::new("darwin", "amd64", "darwin_x64"),
	[BuildTarget]::new("darwin", "amd64", "darwin_x86"),
	[BuildTarget]::new("darwin", "arm64", "darwin_arm64")
)

$jobs = @()

$targets | ForEach-Object {
	$jobs += Start-ThreadJob `
		-Name $_.BuildDir `
		-StreamingHost $Host `
		-ScriptBlock {
			$item = $using:_
			$OS = $item.OS
			$Arch = $item.Arch
			$BuildDir = $item.BuildDir
			$rootDir = $using:rootdir
			$GoExe = $using:GoExe

			Write-Host "Building for $OS $ARCH"

			Set-Location $rootDir\build\$BuildDir

			New-Item -Path Env:\GOOS -Value $OS -Force | Out-Null
			New-Item -Path Env:\GOARCH -Value $ARCH -Force | Out-Null
			& $GoExe build ../..

			Write-Host "Compressing $BuildDir"

			Set-Location $rootDir\build
			7z a -t7z -m0=lzma -mx=9 -mfb=64 -md=32m -ms=on "$BuildDir.7z" (Get-ChildItem .\$BuildDir\*) | Out-Null
			7z a -mx=9 -mfb=64 "$BuildDir.zip" (Get-ChildItem .\$BuildDir\*) | Out-Null
		}
}

Wait-Job -Job $jobs 


mod datetime;
mod fileutil;
mod hookio;
mod platform;
mod sysutil;
#[cfg(test)]
mod testutil;

use std::process::ExitCode;

fn main() -> ExitCode {
    let args: Vec<String> = std::env::args().collect();
    if args.len() < 2 {
        eprintln!("Usage: claudey <subcommand>");
        return ExitCode::from(1);
    }
    eprintln!("Unknown subcommand: {}", args[1]);
    ExitCode::from(1)
}

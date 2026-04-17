use std::process::Command;

#[test]
fn prints_usage_when_no_subcommand() {
    let out = Command::new(env!("CARGO_BIN_EXE_claudey"))
        .output()
        .expect("run binary");
    assert!(!out.status.success());
    let stderr = String::from_utf8_lossy(&out.stderr);
    assert!(stderr.contains("Usage: claudey <subcommand>"));
}

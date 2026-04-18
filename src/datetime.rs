#![allow(dead_code)]

use chrono::Local;

/// Returns the current local date in `YYYY-MM-DD` format.
pub fn date_string() -> String {
    Local::now().format("%Y-%m-%d").to_string()
}

/// Returns the current local time in `HH:MM` format.
pub fn time_string() -> String {
    Local::now().format("%H:%M").to_string()
}

/// Returns the current local datetime in `YYYY-MM-DD HH:MM:SS` format.
pub fn datetime_string() -> String {
    Local::now().format("%Y-%m-%d %H:%M:%S").to_string()
}

#[cfg(test)]
mod tests {
    use super::*;
    use regex::Regex;

    #[test]
    fn test_date_string_format() {
        let s = date_string();
        let re = Regex::new(r"^\d{4}-\d{2}-\d{2}$").unwrap();
        assert!(re.is_match(&s), "date_string() = {s:?}, want YYYY-MM-DD");
    }

    #[test]
    fn test_time_string_format() {
        let s = time_string();
        let re = Regex::new(r"^\d{2}:\d{2}$").unwrap();
        assert!(re.is_match(&s), "time_string() = {s:?}, want HH:MM");
    }

    #[test]
    fn test_datetime_string_format() {
        let s = datetime_string();
        let re = Regex::new(r"^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$").unwrap();
        assert!(
            re.is_match(&s),
            "datetime_string() = {s:?}, want YYYY-MM-DD HH:MM:SS"
        );
    }
}

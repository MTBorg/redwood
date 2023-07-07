use std::env;
use std::path::PathBuf;

use log::LevelFilter;
use log4rs::append::console::ConsoleAppender;
use log4rs::append::file::FileAppender;
use log4rs::config::{Appender, Config, Root};
use log4rs::encode::pattern::PatternEncoder;
use log4rs::filter::threshold::ThresholdFilter;

use crate::cli;

const DEFAULT_LOG_FILE_NAME: &str = ".redwood.log";
const DEFAULT_LOG_FILE_LEVEL: LevelFilter = LevelFilter::Off;
const DEFAULT_STDOUT_LOG_LEVEL: LevelFilter = LevelFilter::Info;

pub struct LogOptions {
    log_file_path: PathBuf,
    log_file_level: LevelFilter,

    stdout_log_level: LevelFilter,
}

impl LogOptions {
    pub fn apply_cli(self, cli: &cli::Cli) -> Self {
        Self {
            log_file_path: cli
                .log_file_path
                .as_ref()
                .unwrap_or(&self.log_file_path)
                .to_path_buf(),
            log_file_level: cli
                .log_file_level
                .as_ref()
                .map(|l| LevelFilter::from(l))
                .unwrap_or(self.log_file_level),

            stdout_log_level: cli
                .log_stdout_level
                .as_ref()
                .map(|l| LevelFilter::from(l))
                .unwrap_or(self.stdout_log_level),
        }
    }
}

impl Default for LogOptions {
    fn default() -> Self {
        LogOptions {
            log_file_path: default_log_path(),
            log_file_level: DEFAULT_LOG_FILE_LEVEL,

            stdout_log_level: DEFAULT_STDOUT_LOG_LEVEL,
        }
    }
}

impl From<&cli::LogFileLevel> for LevelFilter {
    fn from(l: &cli::LogFileLevel) -> Self {
        match l {
            cli::LogFileLevel::Off => Self::Off,
            cli::LogFileLevel::Error => Self::Error,
            cli::LogFileLevel::Warn => Self::Warn,
            cli::LogFileLevel::Info => Self::Info,
            cli::LogFileLevel::Debug => Self::Debug,
            cli::LogFileLevel::Trace => Self::Trace,
        }
    }
}

fn default_log_path() -> std::path::PathBuf {
    let home_dir = match env::var_os("HOME") {
        Some(d) => d,
        None => panic!("HOME environment variable is not existing"),
    };
    return std::path::Path::new(&home_dir).join(DEFAULT_LOG_FILE_NAME);
}

pub fn init(options: LogOptions) {
    let mut root = Root::builder();
    let mut config = Config::builder();

    let stdout = ConsoleAppender::builder().build();
    root = root.appender("stdout");
    config = config.appender(
        Appender::builder()
            .filter(Box::new(ThresholdFilter::new(options.stdout_log_level)))
            .build("stdout", Box::new(stdout)),
    );

    let files = FileAppender::builder()
        .encoder(Box::new(PatternEncoder::new("{d} [{l}] - {m}{n}")))
        .build(options.log_file_path)
        .unwrap();
    root = root.appender("files");
    config = config.appender(
        Appender::builder()
            .filter(Box::new(ThresholdFilter::new(options.log_file_level)))
            .build("files", Box::new(files)),
    );

    const ALL_LEVELS: LevelFilter = LevelFilter::Trace; // let appenders specify level themselves
    let root = root.build(ALL_LEVELS);
    let config = config.build(root).unwrap();

    log4rs::init_config(config).unwrap();
}

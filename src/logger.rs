use std::env;
use std::path::PathBuf;

use log::LevelFilter;
use log4rs::append::console::ConsoleAppender;
use log4rs::append::file::FileAppender;
use log4rs::config::{Appender, Config, Root};
use log4rs::encode::pattern::PatternEncoder;

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

            stdout_log_level: DEFAULT_STDOUT_LOG_LEVEL, // TODO: Parse from cli
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
    let stdout = ConsoleAppender::builder().build();
    let mut root = Root::builder();
    let mut config = Config::builder();

    root = root.appender("stdout");
    config = config.appender(Appender::builder().build("stdout", Box::new(stdout)));

    if options.log_file_level != LevelFilter::Off {
        let files = FileAppender::builder()
            .encoder(Box::new(PatternEncoder::new("{d} - {m}{n}")))
            .build(options.log_file_path)
            .unwrap();

        config = config.appender(Appender::builder().build("files", Box::new(files)));
        root = root.appender("files");
    }

    let root = root.build(options.stdout_log_level); // TODO: Separate loggers for file and stdout
    let config = config.build(root).unwrap();

    log4rs::init_config(config).unwrap();
}

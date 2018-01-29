extern crate clap;
extern crate librespot;
extern crate tokio_core;
//#[macro_use] extern crate configure;
//
//extern crate serde;
//#[macro_use] extern crate serde_derive;

use clap::{Arg, App, SubCommand};

//#[derive(Deserialize, Configure)]
//struct Config {
//    username: String,
//    password: String
//}

use tokio_core::reactor::Core;

use librespot::core::authentication::Credentials;
use librespot::core::config::{PlayerConfig, SessionConfig};
use librespot::core::session::Session;
use librespot::core::util::SpotifyId;

use librespot::audio_backend;
use librespot::player::Player;

fn main() {
    let matches = App::new("Spotify Sync")
        .version("0.1.0")
        .author("Ze'ev Klapow <zklapow@gmail.com>")
        .arg(Arg::with_name("user")
            .short("u")
            .long("user")
            .required(true)
            .help("Spotify User Name")
            .takes_value(true))
        .arg(Arg::with_name("password")
            .short("p")
            .long("password")
            .help("Spotify Password")
            .required(true)
            .takes_value(true))
        .arg(Arg::with_name("track")
            .short("t")
            .long("track")
            .required(true)
            .help("Spotify track ID")
            .takes_value(true))
        .get_matches();

    let username = matches.value_of("user").expect("user is required");
    let password = matches.value_of("password").expect("password is required");
    let track = SpotifyId::from_base62(
        matches.value_of("track")
            .expect("track is required"));

    let mut core = Core::new().unwrap();
    let handle = core.handle();

    let session_config = SessionConfig::default();
    let player_config = PlayerConfig::default();

    let credentials = Credentials::with_password(username.to_owned(), password.to_owned());
    let backend = audio_backend::find(None).unwrap();

    println!("Connecting ..");
    let session = core.run(Session::connect(session_config, credentials, None, handle)).unwrap();

    let player = Player::new(player_config, session.clone(), None, move || (backend)(None));

    println!("Playing track {:?}", track);
    core.run(player.load(track, true, 0)).unwrap();

    println!("Done");
}

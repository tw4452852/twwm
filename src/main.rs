#[macro_use]
extern crate penrose;

use penrose::{
    contrib::extensions::Scratchpad,
    core::{helpers::index_selectors, hooks::Hook},
    draw::{dwm_bar, Color, TextStyle},
    logging_error_handler,
    xcb::{new_xcb_backed_window_manager, XcbDraw},
    Backward, Config, Forward, Less, More, Selector, XcbConnection,
};
use simplelog::{LevelFilter, SimpleLogger};
use std::convert::TryFrom;

const BLACK: &str = "#282828";
const PURPLE: &str = "#d7afff";
const YELLOW: &str = "#ffffd7";
const FONT: &str = "Go Mono";

const TERMINAL: &str = "urxvt";
const CLOCK: &str = "dclock -bg black -led_off black -date '%Y-%b-%d-%a' -bw 0";

fn main() -> penrose::Result<()> {
    SimpleLogger::init(LevelFilter::Info, simplelog::Config::default())
        .expect("failed to init logging");

    let config = Config::default()
        .builder()
        .focused_border(0xd7afff)
        .build()
        .unwrap();

    let sp = Scratchpad::new(CLOCK, 0.8, 0.5);
    let hooks: Vec<Box<dyn Hook<XcbConnection>>> = vec![
        sp.get_hook(),
        Box::new(dwm_bar(
            XcbDraw::new()?,
            18,
            &TextStyle {
                font: FONT.to_string(),
                point_size: 11,
                fg: Color::try_from(BLACK)?,
                bg: Some(Color::try_from(YELLOW)?),
                padding: (2.0, 2.0),
            },
            Color::try_from(PURPLE)?, // highlight
            Color::try_from(YELLOW)?, // empty_ws
            config.workspaces().clone(),
        )?),
    ];

    let key_bindings = gen_keybindings! {
        "M-j" => run_internal!(cycle_client, Forward);
        "M-k" => run_internal!(cycle_client, Backward);
        "M-C-j" => run_internal!(drag_client, Forward);
        "M-C-k" => run_internal!(drag_client, Backward);

        "M-q" => run_internal!(kill_client);
        "M-f" => run_internal!(toggle_client_fullscreen, &Selector::Focused);

        "M-Tab" => run_internal!(toggle_workspace);

        "M-C-Up" => run_internal!(update_max_main, More);
        "M-C-Down" => run_internal!(update_max_main, Less);
        "M-C-Right" => run_internal!(update_main_ratio, More);
        "M-C-Left" => run_internal!(update_main_ratio, Less);

        "M-grave" => run_internal!(cycle_layout, Forward);
        "M-C-grave" => run_internal!(cycle_layout, Backward);

        "M-semicolon" => run_external!("dmenu_run");
        "M-Return" => run_external!(TERMINAL);
        "M-slash" => sp.toggle();
        "M-C-Escape" => run_internal!(exit);

        refmap [ 1..10 ] in {
            "M-{}" => focus_workspace [ index_selectors(9) ];
            "M-C-{}" => client_to_workspace [ index_selectors(9) ];
        };
    };

    let mut wm = new_xcb_backed_window_manager(config, hooks, logging_error_handler())?;
    wm.grab_keys_and_run(key_bindings, map! {})
}

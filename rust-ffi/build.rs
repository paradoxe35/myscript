use std::env;
use std::path::PathBuf;

fn main() {
    let crate_dir = env::var("CARGO_MANIFEST_DIR").unwrap();
    let crate_path = PathBuf::from(&crate_dir);

    cbindgen::Builder::new()
        .with_crate(&crate_dir)
        .with_config(
            cbindgen::Config::from_file(crate_path.join("cbindgen.toml"))
                .expect("Failed to read cbindgen.toml"),
        )
        .generate()
        .expect("Unable to generate C bindings with cbindgen")
        .write_to_file(crate_path.join("bindings.h"));

    println!("cargo:rerun-if-changed=src/");
    println!("cargo:rerun-if-changed=cbindgen.toml");
}

use std::env;
use std::fs::File;
use std::io::Result;
use std::os::fd::AsRawFd;
use std::os::unix::fs::MetadataExt;
use std::slice;

fn main() -> Result<()> {
    let args: Vec<String> = env::args().collect();
    if args.len() != 2 {
        panic!("usage: program <file>\n");
    }

    let file_name = args.get(1).expect("failed to get file name from input");

    let file = File::open(file_name)?;

    let metadata = file.metadata()?;
    if !metadata.is_file() {
        panic!("provided file in not a regular file");
    }

    let data_ptr = unsafe {
        libc::mmap(
            0 as *mut libc::c_void,
            metadata.size() as libc::size_t,
            libc::PROT_READ,
            libc::MAP_SHARED,
            file.as_raw_fd() as libc::c_int,
            0 as libc::off_t,
        )
    };
    if data_ptr == libc::MAP_FAILED {
        panic!("mmap failed");
    }

    drop(file);

    let bytes = unsafe { slice::from_raw_parts(data_ptr as *const u8, metadata.size() as usize) };

    for byte in bytes.iter() {
        print!("{}", *byte as char);
    }

    let munmap_result = unsafe {
        libc::munmap(
            data_ptr as *mut libc::c_void,
            metadata.size() as libc::size_t,
        )
    };
    if munmap_result == -1 {
        panic!("munmap failed");
    }

    Ok(())
}

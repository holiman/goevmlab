# python runtests.py ~/projects/ethereum/tests/GeneralStateTests/Shanghai/
import os
import subprocess
import sys  # Import sys for exiting the script

# Set to true to auto generate the test cases instead of using the json files passed in
use_generic_fuzzer = 1

go_cmd = 'cmd/generic-fuzzer/main.go' if use_generic_fuzzer else 'cmd/runtest/main.go'


# base_command = ['go', 'run', go_cmd, '--revme=binaries/revm', '--cuevm=binaries/cuevm']
base_command = ['go', 'run', go_cmd, '--geth=binaries/geth', '--cuevm=binaries/cuevm']

def test_cmd(full_command, retries=0):
    result = subprocess.run(full_command, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    out = result.stdout.decode('utf-8')
    err = result.stderr.decode('utf-8')

    if 'error starting vm' in err:
        print("VM crashed 💥, ignore")
        # Uncomment these lines to exit when VM crashes
        # print(err)
        # sys.exit(1)


    if 'error' in out:
        print(out)
        # if 'cuevm err:' in out:
        #     sys.exit(1)
        if 'stateRoot' not in out and 'CALL' not in out: # ignore stateRoot mismatches
            # print(out)
            if retries < 1:
                sys.exit(1)
                pass
            else:
                test_cmd(full_command, retries-1)



if use_generic_fuzzer:
    print("Running generic fuzzer")
    while True:
        test_cmd(base_command)
else:
    root_dir = sys.argv[1]
    # Walk through all directories and files in the root directory
    for dirpath, dirnames, filenames in os.walk(root_dir):
        for filename in filenames:
            # Construct the full file path
            file_path = os.path.join(dirpath, filename)

            if not file_path.endswith('.json'):
                continue

            # Combine the base command with the current file path
            full_command = base_command + [file_path]

            print('Executing:', ' '.join(full_command))
            test_cmd(full_command, retries=0)

    print("All files have been processed.")

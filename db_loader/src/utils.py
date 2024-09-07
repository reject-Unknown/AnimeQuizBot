import argparse

def parse_args():
    parser = argparse.ArgumentParser(
                    prog='ProgramName',
                    description='What the program does',
                    epilog='Text at the bottom of help')

    parser.add_argument('-u', '--user', required=True)
    parser.add_argument('-p', '--password', required=True)

    return parser.parse_args()
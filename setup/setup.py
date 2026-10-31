import os
import json
import boto3
import subprocess
from jsonschema import validate, exceptions

# Absolute path to this script directory
SCRIPT_ABS_PATH = os.path.dirname(os.path.abspath(__file__))

# Excepted JSON schema of the urls file
JSON_SCHEMA = {
    "type": "array",
    "items": {
        "type": "string",
        "format": "uri"
    }
}

# --- Helper Functions ---

# Detect and return public IP address
def get_ip():
    print("Detecting workstation public IP address...")
    
    ip_resp = subprocess.run(
        ["curl", "-s", "https://checkip.amazonaws.com"],
        capture_output=True,
        text=True
    )
    
    raw_ip = ip_resp.stdout.strip()
    if not raw_ip:
        print("ERROR: Unable to detect public IP address via checkip.amazonaws.com")
        exit(1)
        
    return f"{raw_ip}/32"

# Generate SSH public key to register in AWS
def gen_pubkey(private_key_path):
    print("Generating an SSH public key for an AWS EC2 instance...")
    
    try:
        pub_key = subprocess.check_output(
            ["ssh-keygen", "-y", "-f", private_key_path],
            stderr=subprocess.DEVNULL
        )
        return pub_key.decode("utf-8").strip()
    except subprocess.CalledProcessError:
        print(f"ERROR: Unable to extract public key from {private_key_path}")
        return None
    
    

# Parse the json file content and validate it.
def is_valid_json(file):
    print("Validating the JSON file structure...")
    
    try:
        with open(file) as f:
            data = json.loads(f.read())
            validate(data, JSON_SCHEMA)
            return True
    except exceptions.ValidationError as e:
        print(f"ERROR validating the .json file: {e.message}")
        return False
    except Exception as e:
        print(f"ERROR reading file: {e}")
        return False
        
# Run 'terraform apply' to set up the cloud infrastructure.
def apply_terraform_config():
    print("Applying Terraform configuration and setting up the cloud infrastructure...")
    terraform_dir = os.path.abspath(os.path.join(SCRIPT_ABS_PATH, "../terraform"))
    
    process = subprocess.Popen(
        ["terraform", "apply", "-no-color", "-auto-approve"],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        bufsize=1,
        cwd=terraform_dir,
        text=True
    )

    for line in process.stdout:
        print(line, end="")
        
    return process.wait()

# Upload the json urls file to the S3 bucket
def upload_urls_to_s3(file_name, bucket_name):
    s3_object_name = os.path.basename(file_name)
    print(f"Uploading {s3_object_name} to S3 bucket '{bucket_name}'...")
    
    s3 = boto3.client('s3')
    s3.upload_file(file_name, bucket_name, s3_object_name)
    print("Upload complete!")
    
def trigger_initial_deployment():
    print("Attempting to trigger initial deployment via GitHub Actions...")
    
    # Try using gh CLI first
    try:
        result = subprocess.run(
            ["gh", "workflow", "run", "deploy.yml"],
            capture_output=True,
            text=True
        )
        if result.returncode == 0:
            print("Initial deployment triggered successfully!")
            return
    except FileNotFoundError:
        # gh CLI is not installed on the system
        pass

    # Fallback message if gh is missing, unauthenticated, or fails
    print("\n" + "="*60)
    print("NOTE: Could not automatically trigger GitHub Actions deployment.")
    print("To launch your application:")
    print("  1. Push a new commit to your repository, OR")
    print("  2. Run 'gh auth login' and then 'gh workflow run deploy.yml', OR")
    print("  3. Trigger the 'deploy.yml' workflow manually from the GitHub Actions tab.")
    print("="*60 + "\n")
    
# --- Main Entry Point ---

def main():
    # Set ~ as a user home directory
    os_home = os.path.expanduser("~")

    # User inputs
    private_key_path = str(input("Enter your local SSH private key (default: ~/.ssh/id_rsa): ").strip() or f"{os_home}/.ssh/id_rsa")
    key_pair_name = str(input("Enter your desired key name label to register in AWS (default: site-checker-key): ").strip() or "site-checker-key")
    bucket_name = str(input("Enter your AWS S3 Bucket name (default: urls-bucket-1928): ").strip() or "urls-bucket-1928")
    local_urls_path = str(input("Enter the path to the .json file with your urls (default: urls.json): ").strip() or "urls.json")
    print("")
    
    # Expands ~ if the user typed it explicitly in the terminal
    private_key_path = os.path.expanduser(private_key_path)
    local_urls_path  = os.path.expanduser(local_urls_path)

    # Check if the SSH private key path is valid
    if not os.path.isfile(private_key_path):
        print(f"ERROR: {private_key_path} doesn't exist.")
        exit(1)

    # Check if the urls file path is valid
    if not os.path.isfile(local_urls_path):
        print(f"ERROR: {local_urls_path} doesn't exist.")
        exit(1)

    # Check if the json structure is valid
    if not is_valid_json(local_urls_path):
        exit(1)

    # Get public IP address
    ip_addr = get_ip()

    # Extract public key
    pub_key = gen_pubkey(private_key_path)
    if not pub_key:
        exit(1)

    # Write to terraform.tfvars
    tfvars_path = os.path.abspath(os.path.join(SCRIPT_ABS_PATH, "../terraform/terraform.tfvars"))
    print(f"Updating {tfvars_path}...")

    with open(tfvars_path, "w") as f:
        f.write(
            f'# terraform/terraform.tfvars\n'
            f'my_ip             = "{ip_addr}"\n'
            f'bucket_name       = "{bucket_name}"\n'
            f'key_name          = "{key_pair_name}"\n'
            f'public_key_string = "{pub_key}"\n'
        )
        
    # Apply Terraform configuration
    return_code = apply_terraform_config()
    if return_code != 0:
        print(f"\nTerraform failed with exit code {return_code}")
        exit(1)
        
    # Upload JSON to S3
    upload_urls_to_s3(local_urls_path, bucket_name)
    
    # Deploy the initial Docker image to the EC2 instance
    trigger_initial_deployment()
    

if __name__ == '__main__':
    main()
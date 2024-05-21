require 'vagrant-env'
require 'vagrant-openstack-provider'

Vagrant.configure('2') do |config|
  config.env.enable
  config.ssh.username = ENV['OS_SSH_USERNAME']

  num_instances = 6
  project_name = "ssbench"

  (1..num_instances).each do |i|
    config.vm.define "#{project_name}#{i}" do |vm_config|

      vm_config.vm.provider :openstack do |os|
        os.identity_api_version             = ENV['OS_IDENTITY_API_VERSION']
        os.openstack_auth_url               = ENV['OS_AUTH_URL']
        os.project_name                     = ENV['OS_PROJECT_NAME']
        os.user_domain_name                 = ENV['OS_USER_DOMAIN_NAME']
        os.project_domain_name              = ENV['OS_PROJECT_DOMAIN_ID']
        os.username                         = ENV['OS_USERNAME']
        os.password                         = ENV['OS_PASSWORD']
        os.region                           = ENV['OS_REGION_NAME']
        os.flavor                           = ENV['OS_FLAVOR']
        os.image                            = ENV['OS_IMAGE']
        os.interface_type                   = ENV['OS_INTERFACE']
        os.availability_zone                = ENV['OS_AVAILABILITY_ZONE']

        os.server_create_timeout            = 3600
        os.server_delete_timeout            = 3600
      end
    end
  end
end

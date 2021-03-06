package gitlab

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/xanzy/go-gitlab"
)

func resourceGitlabGroupVariable() *schema.Resource {
	return &schema.Resource{
		Create: resourceGitlabGroupVariableCreate,
		Read:   resourceGitlabGroupVariableRead,
		Update: resourceGitlabGroupVariableUpdate,
		Delete: resourceGitlabGroupVariableDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"group": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"key": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: StringIsGitlabVariableName(),
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"protected": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceGitlabGroupVariableCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	group := d.Get("group").(string)
	key := d.Get("key").(string)
	value := d.Get("value").(string)
	protected := d.Get("protected").(bool)

	options := gitlab.CreateVariableOptions{
		Key:              &key,
		Value:            &value,
		Protected:        &protected,
		EnvironmentScope: nil,
	}
	log.Printf("[DEBUG] create gitlab group variable %s/%s", group, key)

	_, _, err := client.GroupVariables.CreateVariable(group, &options)
	if err != nil {
		return err
	}

	d.SetId(buildTwoPartID(&group, &key))

	return resourceGitlabGroupVariableRead(d, meta)
}

func resourceGitlabGroupVariableRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	group, key, err := parseTwoPartID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] read gitlab group variable %s/%s", group, key)

	v, resp, err := client.GroupVariables.GetVariable(group, key)
	if err != nil {
		if resp.StatusCode == 404 {
			log.Printf("[WARN] Group variable %v for group %s not found in gitlab", key, group)
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("key", v.Key)
	d.Set("value", v.Value)
	d.Set("group", group)
	d.Set("protected", v.Protected)
	return nil
}

func resourceGitlabGroupVariableUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)

	group := d.Get("group").(string)
	key := d.Get("key").(string)
	value := d.Get("value").(string)
	protected := d.Get("protected").(bool)

	options := &gitlab.UpdateVariableOptions{
		Key:              &key,
		Value:            &value,
		Protected:        &protected,
		EnvironmentScope: nil,
	}
	log.Printf("[DEBUG] update gitlab group variable %s/%s", group, key)

	_, resp, err := client.GroupVariables.UpdateVariable(group, key, options)
	if err != nil {
		if resp.StatusCode == 404 {
			log.Printf("[WARN] Group variable %v for group %s not found in gitlab, removing from state", key, group)
			d.SetId("")
			return nil
		}
		return err
	}
	return resourceGitlabGroupVariableRead(d, meta)
}

func resourceGitlabGroupVariableDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*gitlab.Client)
	group := d.Get("group").(string)
	key := d.Get("key").(string)
	log.Printf("[DEBUG] Delete gitlab group variable %s/%s", group, key)

	_, err := client.GroupVariables.RemoveVariable(group, key)
	return err
}

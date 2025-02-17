package emr

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_emr_studio_session_mapping")
func ResourceStudioSessionMapping() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStudioSessionMappingCreate,
		ReadWithoutTimeout:   resourceStudioSessionMappingRead,
		UpdateWithoutTimeout: resourceStudioSessionMappingUpdate,
		DeleteWithoutTimeout: resourceStudioSessionMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"identity_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ExactlyOneOf: []string{"identity_id", "identity_name"},
			},
			"identity_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Computed:     true,
				ExactlyOneOf: []string{"identity_id", "identity_name"},
			},
			"identity_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(emr.IdentityType_Values(), false),
			},
			"session_policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"studio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceStudioSessionMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn()

	var id string
	studioId := d.Get("studio_id").(string)
	identityType := d.Get("identity_type").(string)
	input := &emr.CreateStudioSessionMappingInput{
		IdentityType:     aws.String(identityType),
		SessionPolicyArn: aws.String(d.Get("session_policy_arn").(string)),
		StudioId:         aws.String(studioId),
	}

	if v, ok := d.GetOk("identity_id"); ok {
		input.IdentityId = aws.String(v.(string))
		id = v.(string)
	}

	if v, ok := d.GetOk("identity_name"); ok {
		input.IdentityName = aws.String(v.(string))
		id = v.(string)
	}

	_, err := conn.CreateStudioSessionMappingWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Studio Session Mapping: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s:%s", studioId, identityType, id))

	return append(diags, resourceStudioSessionMappingRead(ctx, d, meta)...)
}

func resourceStudioSessionMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn()

	studioId, identityType, identityId, err := readStudioSessionMapping(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EMR Studio Session Mapping (%s): %s", d.Id(), err)
	}

	input := &emr.UpdateStudioSessionMappingInput{
		SessionPolicyArn: aws.String(d.Get("session_policy_arn").(string)),
		IdentityType:     aws.String(identityType),
		StudioId:         aws.String(studioId),
		IdentityId:       aws.String(identityId),
	}

	_, err = conn.UpdateStudioSessionMappingWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EMR Studio Session Mapping (%s): %s", d.Id(), err)
	}

	return append(diags, resourceStudioSessionMappingRead(ctx, d, meta)...)
}

func resourceStudioSessionMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn()

	mapping, err := FindStudioSessionMappingByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Studio Session Mapping (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Studio Session Mapping (%s): %s", d.Id(), err)
	}

	d.Set("identity_type", mapping.IdentityType)
	d.Set("identity_id", mapping.IdentityId)
	d.Set("identity_name", mapping.IdentityName)
	d.Set("studio_id", mapping.StudioId)
	d.Set("session_policy_arn", mapping.SessionPolicyArn)

	return diags
}

func resourceStudioSessionMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn()
	studioId, identityType, identityId, err := readStudioSessionMapping(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EMR Studio Session Mapping (%s): %s", d.Id(), err)
	}

	input := &emr.DeleteStudioSessionMappingInput{
		IdentityType: aws.String(identityType),
		StudioId:     aws.String(studioId),
		IdentityId:   aws.String(identityId),
	}

	log.Printf("[INFO] Deleting EMR Studio Session Mapping: %s", d.Id())
	_, err = conn.DeleteStudioSessionMappingWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "Studio session mapping does not exist.") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting EMR Studio Session Mapping (%s): %s", d.Id(), err)
	}

	return diags
}

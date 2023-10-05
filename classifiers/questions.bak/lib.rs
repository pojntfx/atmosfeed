use signature::types;

pub fn scale(
    ctx: Option<&mut types::Context>,
) -> Result<Option<types::Context>, Box<dyn std::error::Error>> {
    let mut changed_ctx = ctx.map(|c| c.clone());

    if let Some(context) = changed_ctx.as_mut() {
        context.weight = context.post.as_ref().map_or(-1, |post| {
            if post.text.contains("?") {
                post.likes
            } else {
                -1
            }
        });
    }

    signature::next(changed_ctx)
}

local function handle()
    local data = ctx.form_data

	local id, error = storage_save("orders", data)

	local priority_tbl =
	{
	  ["high"] = "Важнейший заказ: \n",
	  ["medium"] = "Заказ: \n",
	  ["low"] = "Можно сделать: \n",
	}
	
	send_channel(priority_tbl[data.priority] .. data.info, {
		Rows = {
		  {
			{Text = "Принять", Data = ctx.user.id .. "|" .. id, Script = "takeorder.lua"},
		  }
		}})

end

handle()
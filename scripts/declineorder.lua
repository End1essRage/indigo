local function handle()
    local order_id = ctx.cb_data.data

	local order, error = storage_load("orders", order_id)

	local priority_tbl =
	{
	  ["high"] = "Важнейший заказ: \n",
	  ["medium"] = "Заказ: \n",
	  ["low"] = "Можно сделать: \n",
	}

	send_channel(priority_tbl[order.priority] .. order.info, {
		Rows = {
		  {
			{Text = "Принять", Data = ctx.user.id .. "|" .. order_id, Script = "takeorder.lua"},
		  }
		}})

end

handle()
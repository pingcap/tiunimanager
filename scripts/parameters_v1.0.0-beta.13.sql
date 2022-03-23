update parameter_group_mappings set default_value = '' where default_value = ' ';
update parameter_group_mappings set default_value = '' where parameter_id in ('62', '74', '86', '98', '110', '111', '117', '186', '209', '240', '271');
delete from parameter_group_mappings where parameter_group_id = 'II7PQbJWTHW2_VdlNRjtrQ' and parameter_id = '187';
delete from parameter_group_mappings where parameter_group_id = 'II7PQbJWTHW2_VdlNRjtrQ' and parameter_id = '190';
delete from parameter_group_mappings where parameter_group_id = 'SqsfgL4tQ7GaBGOxIK1c8A' and parameter_id = '187';
delete from parameter_group_mappings where parameter_group_id = 'SqsfgL4tQ7GaBGOxIK1c8A' and parameter_id = '188';
delete from parameter_group_mappings where parameter_group_id = 'SqsfgL4tQ7GaBGOxIK1c8A' and parameter_id = '189';
delete from parameter_group_mappings where parameter_group_id = 'SqsfgL4tQ7GaBGOxIK1c8A' and parameter_id = '190';
delete from parameter_group_mappings where parameter_group_id = 'SqsfgL4tQ7GaBGOxIK1c8A' and parameter_id = '191';